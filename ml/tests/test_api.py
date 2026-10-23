from fastapi.testclient import TestClient

from app.main import app


client = TestClient(app)


def test_fixed_company_contracts():
    inn = "7701234560"

    egrul = client.get(f"/egrul/{inn}").json()
    fns = client.get(f"/fns/{inn}").json()
    bki = client.get(f"/bki/{inn}").json()
    accounts = client.get(f"/abs/{inn}/accounts").json()

    assert egrul["name"] == "ООО «Колос и Ко»"
    assert egrul["directorName"] == "Иванова Ирина Сергеевна"
    assert len(fns["monthlyRevenue"]) == 12
    assert fns["monthlyRevenue"][0]["month"] == "2025-10"
    assert fns["monthlyRevenue"][-1]["month"] == "2026-09"
    assert fns["accountsBlocked"] is False
    assert sum(o["balance"] for o in bki["obligations"]) == 4_100_000
    assert round(sum(o["monthlyPayment"] for o in bki["obligations"])) == 110_000
    assert accounts["totalBalance"] == 1_284_300
    assert accounts["accounts"][0]["number"] == "40702810400000012345"


def test_liquidation_stop_factor():
    response = client.get("/egrul/5251234564")

    assert response.status_code == 200
    assert response.json()["status"] == "LIQUIDATING"
    assert response.json()["stopFactor"] == "LIQUIDATION"


def test_mock_fault_header():
    assert client.get("/bki/7701234560", headers={"X-Mock-Fault": "timeout"}).status_code == 504
    assert client.get("/bki/7701234560", headers={"X-Mock-Fault": "error"}).status_code == 503


def test_disbursement():
    response = client.post(
        "/abs/disbursements",
        json={"contractNumber": "KD-1", "accountNumber": "40702810400000012345", "amount": 1_000_000},
    )

    assert response.status_code == 201
    assert response.json()["operationId"].startswith("op-")


def test_disbursement_is_idempotent():
    body = {"contractNumber": "KD-2", "accountNumber": "40702810400000012345", "amount": 500_000}

    first = client.post("/abs/disbursements", json=body).json()
    second = client.post("/abs/disbursements", json=body).json()

    assert first == second


def test_fns_demo_failure():
    response = client.get("/fns/7700000503")

    assert response.status_code == 503


def test_unknown_inn_is_stable():
    path = "/fns/123456789012"

    assert client.get(path).json() == client.get(path).json()


def test_invalid_inn_is_rejected():
    assert client.get("/bki/not-an-inn").status_code == 422


def test_recommend_falls_back_to_rules_without_llm(monkeypatch):
    for name in ("OLLAMA_URL", "LLM_API_URL", "LLM_API_KEY"):
        monkeypatch.delenv(name, raising=False)

    response = client.post(
        "/llm/recommend",
        json={"purpose": "Закупка товара к сезону", "amount": 2_000_000, "termMonths": 12},
    )

    assert response.status_code == 200
    assert response.json()["productCode"] == "WORKING_CAPITAL"
    assert response.json()["source"] == "RULES"


def test_explain_falls_back_to_rules_without_llm(monkeypatch):
    for name in ("OLLAMA_URL", "LLM_API_URL", "LLM_API_KEY"):
        monkeypatch.delenv(name, raising=False)

    response = client.post(
        "/llm/explain",
        json={
            "outcome": "APPROVED",
            "score": 742,
            "pd": 0.034,
            "pGap": 0.12,
            "factors": ["стабильная выручка", "нет просрочек"],
            "pointA": "ENRICHING",
            "pointB": "APPROVED",
        },
    )

    assert response.status_code == 200
    assert response.json()["source"] == "RULES"
    assert "3.4%" in response.json()["decisionReason"]


def test_valid_llm_response_is_used(monkeypatch):
    async def fake_llm(_prompt):
        return {"productCode": "CUSTOM", "explanation": "Ответ модели"}

    monkeypatch.setattr("app.llm._request_llm", fake_llm)
    response = client.post(
        "/llm/recommend",
        json={"purpose": "Особая цель", "amount": 100_000, "termMonths": 6},
    )

    assert response.json() == {
        "productCode": "CUSTOM",
        "explanation": "Ответ модели",
        "source": "LLM",
    }


def test_slow_llm_falls_back_to_rules(monkeypatch):
    import asyncio

    async def slow_llm(_prompt):
        await asyncio.sleep(0.05)
        return {"productCode": "TOO_LATE", "explanation": "Поздний ответ"}

    monkeypatch.setattr("app.llm.TIMEOUT_SECONDS", 0.01)
    monkeypatch.setattr("app.llm._request_llm", slow_llm)
    response = client.post(
        "/llm/recommend",
        json={"purpose": "Закупка товара", "amount": 100_000, "termMonths": 6},
    )

    assert response.json()["source"] == "RULES"

