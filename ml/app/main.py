import hashlib
import json
import random
import re
from datetime import date, datetime, timezone
from pathlib import Path

from fastapi import FastAPI, Header, HTTPException
from pydantic import BaseModel, Field

from app.llm import ExplainRequest, ExplainResponse, RecommendRequest, RecommendResponse
from app.llm import explain as explain_decision
from app.llm import recommend as recommend_product


app = FastAPI(title="ML Mock Service", version="1.0.0")
COMPANIES = json.loads(Path(__file__).with_name("companies.json").read_text(encoding="utf-8"))
FAILED_FNS_INN = "7700000503"
INN_PATTERN = re.compile(r"^(?:\d{10}|\d{12})$")
NAMES = ("ООО «Вектор»", "ООО «Маяк»", "ООО «Горизонт»", "ООО «Партнёр»")
DIRECTORS = (
    "Соколов Михаил Петрович",
    "Петрова Анна Владимировна",
    "Васильев Сергей Ильич",
    "Морозова Ольга Алексеевна",
)
OKVEDS = ("46.90", "47.19", "62.01", "68.20", "70.22")
OBLIGATION_TYPES = ("LOAN", "LEASING", "OVERDRAFT", "CREDIT_CARD")
EGRUL_STATUS_BY_STOP_FACTOR = {"LIQUIDATION": "LIQUIDATING", "BANKRUPTCY": "BANKRUPTCY"}
TAX_SHARE = 0.06  # упрощённо: УСН 6% от выручки
DISBURSEMENTS: dict[str, dict] = {}  # contractNumber → ответ, живёт до рестарта мока


def _rng(inn: str) -> random.Random:
    return random.Random(int(hashlib.sha256(inn.encode()).hexdigest(), 16))


def _validate_inn(inn: str) -> None:
    if not INN_PATTERN.fullmatch(inn):
        raise HTTPException(status_code=422, detail="INN must contain 10 or 12 digits")


def _generated_company(inn: str) -> dict:
    rng = _rng(inn)
    annual_revenue = rng.randrange(8_000_000, 180_000_001, 100_000)
    return {
        "egrul": {
            "ogrn": "1" + "".join(str(rng.randrange(10)) for _ in range(12 if len(inn) == 10 else 14)),
            "name": NAMES[rng.randrange(len(NAMES))],
            "legalForm": "LLC",
            "okved": OKVEDS[rng.randrange(len(OKVEDS))],
            "registeredAt": date(rng.randint(2004, 2024), rng.randint(1, 12), rng.randint(1, 28)).isoformat(),
            "status": "ACTIVE",
            "director": DIRECTORS[rng.randrange(len(DIRECTORS))],
            "stopFactor": None,
        },
        "fns": {
            "annualRevenue": annual_revenue,
            "employees": rng.randint(3, 120),
            "taxDebt": rng.randrange(0, 400_001, 100),
            "revenueTrend": round(rng.uniform(-0.15, 0.35), 2),
        },
        "bki": {
            "activeLoansCount": rng.randint(0, 5),
            "totalDebtBalance": rng.randrange(0, 20_000_001, 10_000),
            "overdues12m": rng.randint(0, 3),
            "maxOverdueDays": rng.choice((0, 0, 0, 7, 14, 30)),
            "inquiries3m": rng.randint(0, 5),
        },
        "abs": {
            "settlementAccount": "40702810" + "".join(str(rng.randrange(10)) for _ in range(12)),
            "cashBalance": rng.randrange(50_000, 10_000_001, 1_000),
            "currentDebtPayments": rng.randrange(0, 800_001, 1_000),
        },
    }


def _company(inn: str) -> dict:
    _validate_inn(inn)
    return COMPANIES.get(inn) or _generated_company(inn)


def _source(inn: str, source: str) -> dict:
    return {"inn": inn, **_company(inn)[source]}


def _fault(fault: str | None) -> None:
    """X-Mock-Fault из mocks-openapi.yaml: имитация таймаута или ошибки источника."""
    if fault == "timeout":
        raise HTTPException(status_code=504, detail="Source timeout (X-Mock-Fault)")
    if fault == "error":
        raise HTTPException(status_code=503, detail="Source error (X-Mock-Fault)")


def _monthly_revenue(inn: str, annual_revenue: int) -> list[dict]:
    rng = _rng(f"monthly:{inn}")
    average = annual_revenue / 12
    months = []
    year, month = 2025, 10
    for _ in range(12):
        amount = round(average * rng.uniform(0.82, 1.18) / 1000) * 1000
        months.append({"month": f"{year:04d}-{month:02d}", "amount": amount})
        month += 1
        if month == 13:
            year, month = year + 1, 1
    return months


def _egrul_status(egrul: dict) -> str:
    return EGRUL_STATUS_BY_STOP_FACTOR.get(egrul.get("stopFactor"), egrul["status"])


def _obligations(bki: dict, monthly_payments: int) -> list[dict]:
    """Делит общий долг и платёж поровну между активными кредитами; сумма monthlyPayment = D_old."""
    count = bki["activeLoansCount"]
    if count == 0:
        return []
    return [
        {
            "type": OBLIGATION_TYPES[i % len(OBLIGATION_TYPES)],
            "balance": round(bki["totalDebtBalance"] / count),
            "monthlyPayment": round(monthly_payments / count, 2),
        }
        for i in range(count)
    ]


@app.get("/health")
def health() -> dict:
    return {"status": "ok"}


@app.get("/egrul/{inn}")
def egrul(inn: str, x_mock_fault: str | None = Header(default=None)) -> dict:
    _fault(x_mock_fault)
    data = _company(inn)["egrul"]
    return {
        "inn": inn,
        "ogrn": data["ogrn"],
        "name": data["name"],
        "legalForm": data["legalForm"],
        "status": _egrul_status(data),
        "registeredAt": data["registeredAt"],
        "okved": data["okved"],
        "directorName": data["director"],
        "stopFactor": data.get("stopFactor"),
    }


@app.get("/fns/{inn}")
def fns(inn: str, x_mock_fault: str | None = Header(default=None)) -> dict:
    _validate_inn(inn)
    _fault(x_mock_fault)
    if inn == FAILED_FNS_INN:
        raise HTTPException(status_code=503, detail="FNS service temporarily unavailable")
    data = _company(inn)["fns"]
    return {
        "inn": inn,
        "monthlyRevenue": _monthly_revenue(inn, data["annualRevenue"]),
        "taxesPaidLastYear": round(data["annualRevenue"] * TAX_SHARE),
        "taxDebt": data["taxDebt"],
        "accountsBlocked": data.get("accountsBlocked", False),
        "employees": data["employees"],
    }


@app.get("/bki/{inn}")
def bki(inn: str, x_mock_fault: str | None = Header(default=None)) -> dict:
    _fault(x_mock_fault)
    company = _company(inn)
    data = company["bki"]
    return {
        "inn": inn,
        "overdues12m": data["overdues12m"],
        "maxOverdueDays": data["maxOverdueDays"],
        "inquiries3m": data["inquiries3m"],
        "obligations": _obligations(data, company["abs"]["currentDebtPayments"]),
    }


@app.get("/abs/{inn}/accounts")
def accounts(inn: str, x_mock_fault: str | None = Header(default=None)) -> dict:
    _fault(x_mock_fault)
    data = _company(inn)["abs"]
    return {
        "inn": inn,
        "totalBalance": data["cashBalance"],
        "accounts": [{"number": data["settlementAccount"], "balance": data["cashBalance"]}],
    }


class Disbursement(BaseModel):
    contractNumber: str = Field(min_length=1)
    accountNumber: str = ""
    amount: float = Field(gt=0)


@app.post("/abs/disbursements", status_code=201)
def disburse(request: Disbursement, x_mock_fault: str | None = Header(default=None)) -> dict:
    """Идемпотентно по contractNumber: повтор (ретрай конвейера) возвращает ту же операцию и то же время."""
    _fault(x_mock_fault)
    if request.contractNumber not in DISBURSEMENTS:
        operation = hashlib.sha256(request.contractNumber.encode()).hexdigest()[:16]
        DISBURSEMENTS[request.contractNumber] = {
            "operationId": f"op-{operation}",
            "disbursedAt": datetime.now(timezone.utc).isoformat(),
        }
    return DISBURSEMENTS[request.contractNumber]


@app.post("/llm/recommend", response_model=RecommendResponse)
async def llm_recommend(request: RecommendRequest) -> RecommendResponse:
    return await recommend_product(request)


@app.post("/llm/explain", response_model=ExplainResponse)
async def llm_explain(request: ExplainRequest) -> ExplainResponse:
    return await explain_decision(request)

