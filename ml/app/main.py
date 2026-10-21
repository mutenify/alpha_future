import hashlib
import json
import random
import re
from datetime import date
from pathlib import Path

from fastapi import FastAPI, HTTPException

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


@app.get("/health")
def health() -> dict:
    return {"status": "ok"}


@app.get("/egrul/{inn}")
def egrul(inn: str) -> dict:
    return _source(inn, "egrul")


@app.get("/fns/{inn}")
def fns(inn: str) -> dict:
    _validate_inn(inn)
    if inn == FAILED_FNS_INN:
        raise HTTPException(status_code=503, detail="FNS service temporarily unavailable")
    return _source(inn, "fns")


@app.get("/bki/{inn}")
def bki(inn: str) -> dict:
    return _source(inn, "bki")


@app.get("/abs/{inn}/accounts")
def accounts(inn: str) -> dict:
    company = _company(inn)
    return {
        **company["abs"],
        "monthlyRevenue": _monthly_revenue(inn, company["fns"]["annualRevenue"]),
    }


@app.post("/llm/recommend", response_model=RecommendResponse)
async def llm_recommend(request: RecommendRequest) -> RecommendResponse:
    return await recommend_product(request)


@app.post("/llm/explain", response_model=ExplainResponse)
async def llm_explain(request: ExplainRequest) -> ExplainResponse:
    return await explain_decision(request)

