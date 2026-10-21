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
    assert fns["annualRevenue"] == 38_000_000
    assert bki["totalDebtBalance"] == 4_100_000
    assert len(accounts["monthlyRevenue"]) == 12
    assert accounts["monthlyRevenue"][0]["month"] == "2025-10"
    assert accounts["monthlyRevenue"][-1]["month"] == "2026-09"


def test_liquidation_stop_factor():
    response = client.get("/egrul/5251234564")

    assert response.status_code == 200
    assert response.json()["stopFactor"] == "LIQUIDATION"


def test_fns_demo_failure():
    response = client.get("/fns/7700000503")

    assert response.status_code == 503


def test_unknown_inn_is_stable():
    path = "/fns/123456789012"

    assert client.get(path).json() == client.get(path).json()


def test_invalid_inn_is_rejected():
    assert client.get("/bki/not-an-inn").status_code == 422

