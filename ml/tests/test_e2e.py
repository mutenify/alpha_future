from __future__ import annotations

import csv
import os
import time
import uuid
from pathlib import Path

import httpx
import pytest


BASE_URL = os.getenv("E2E_BASE_URL", "http://localhost:8080/api/v1")
FINAL_STATUSES = {"APPROVED", "COUNTER_OFFER", "MANUAL_REVIEW", "REJECTED", "INFO_REQUESTED"}
PD_TOLERANCE = 5.1e-4       # database numeric(5,4)
PGAP_TOLERANCE = 5.1e-4     # database numeric(5,4)
MONEY_TOLERANCE = 0.02      # database numeric(15,2)
RATE_TOLERANCE = 0.011      # database numeric(5,2)


def _golden_cases() -> list[dict[str, str]]:
    path = Path(__file__).resolve().parents[2] / "backend/internal/scoring/testdata/golden_cases.csv"
    with path.open(encoding="utf-8") as source:
        return list(csv.DictReader(source))


def _num(case: dict[str, str], field: str) -> float:
    return float(case[field])


def _assert_close(case: dict[str, str], field: str, actual: float, expected: float, tolerance: float) -> None:
    delta = abs(actual - expected)
    assert delta <= tolerance, (
        f"persona={case['case_id']} field={field} expected={expected:.12g} "
        f"actual={actual:.12g} abs_delta={delta:.12g} tolerance={tolerance:.12g}"
    )


def _login(client: httpx.Client, login: str) -> str:
    response = client.post("/auth/login", json={"login": login, "password": "demo"})
    assert response.status_code == 200, f"login={login}: {response.status_code} {response.text}"
    return response.json()["accessToken"]


def _headers(token: str, key: str | None = None) -> dict[str, str]:
    result = {"Authorization": f"Bearer {token}"}
    if key:
        result["Idempotency-Key"] = key
    return result


@pytest.fixture(scope="module")
def live_client() -> httpx.Client:
    client = httpx.Client(base_url=BASE_URL, timeout=5)
    try:
        response = client.get("/health")
    except httpx.HTTPError as exc:
        pytest.fail(f"live Go API is unavailable at {BASE_URL}: {exc}")
    if response.status_code != 200:
        pytest.fail(f"live Go API health failed: {response.status_code} {response.text}")
    yield client
    client.close()


@pytest.fixture(scope="module")
def inspector_token(live_client: httpx.Client) -> str:
    return _login(live_client, "inspector1")


def _poll(client: httpx.Client, app_id: str, token: str, expected: str, timeout: float = 25) -> dict:
    deadline, seen = time.monotonic() + timeout, []
    while time.monotonic() < deadline:
        response = client.get(f"/applications/{app_id}", headers=_headers(token))
        assert response.status_code == 200, response.text
        application = response.json()
        seen.append(application["status"])
        if application["status"] in FINAL_STATUSES:
            assert application["status"] == expected, f"app={app_id} expected={expected} actual={application['status']} transitions={seen}"
            return application
        time.sleep(0.35)
    pytest.fail(f"app={app_id} polling timeout expected={expected} transitions={seen}")


@pytest.mark.e2e
@pytest.mark.parametrize("case", _golden_cases(), ids=lambda case: f"persona-{case['case_id']}")
def test_live_credit_pipeline(live_client: httpx.Client, inspector_token: str, case: dict[str, str]) -> None:
    token = _login(live_client, case["login"])
    body = {
        "companyInn": case["inn"], "productCode": case["product_code"], "amount": _num(case, "amount"),
        "termMonths": int(case["term_months"]), "purpose": case["purpose"],
    }
    create_key = f"e2e-create-{case['case_id']}-{uuid.uuid4()}"
    created = live_client.post("/applications", headers=_headers(token, create_key), json=body)
    assert created.status_code == 201, f"persona={case['case_id']} create: {created.status_code} {created.text}"
    assert created.json()["status"] == "DRAFT"

    replay = live_client.post("/applications", headers=_headers(token, create_key), json=body)
    assert replay.status_code == 201 and replay.headers.get("Idempotent-Replay") == "true"
    assert replay.json()["id"] == created.json()["id"]

    app_id = created.json()["id"]
    submitted = live_client.post(f"/applications/{app_id}/submit", headers=_headers(token, f"e2e-submit-{uuid.uuid4()}"))
    assert submitted.status_code == 202, f"persona={case['case_id']} submit: {submitted.status_code} {submitted.text}"
    assert submitted.json()["status"] in {"SUBMITTED", "ENRICHING", "SCORING", case["expected_status"]}
    application = _poll(live_client, app_id, token, case["expected_status"])

    offers = application.get("offers") or []
    if case["expected_status"] == "APPROVED":
        assert [offer["variant"] for offer in offers] == ["REQUESTED"]
        _assert_close(case, "offer.amount", offers[0]["amount"], _num(case, "amount"), MONEY_TOLERANCE)
    elif case["expected_status"] == "COUNTER_OFFER":
        expected_variants = {name for name, field in (("POINT_A", "point_a_amount"), ("POINT_B", "point_b_amount")) if case[field]}
        assert {offer["variant"] for offer in offers} == expected_variants
        for offer in offers:
            prefix = "point_a" if offer["variant"] == "POINT_A" else "point_b"
            _assert_close(case, prefix + ".amount", offer["amount"], _num(case, prefix + "_amount"), MONEY_TOLERANCE)
            assert offer["termMonths"] == int(case[prefix + "_term"])
            _assert_close(case, prefix + ".rate", offer["rate"], _num(case, "rate"), RATE_TOLERANCE)
            _assert_close(case, prefix + ".payment", offer["monthlyPayment"], _num(case, prefix + "_payment"), MONEY_TOLERANCE)
            _assert_close(case, prefix + ".pGap", offer["pGap"], _num(case, prefix + "_p_gap"), PGAP_TOLERANCE)
    else:
        assert offers == [], f"persona={case['case_id']} invalid offers for {case['expected_status']}: {offers}"

    scoring = live_client.get(f"/applications/{app_id}/scoring", headers=_headers(inspector_token))
    if case["special"]:
        assert scoring.status_code == 404, f"persona={case['case_id']} special={case['special']} unexpectedly has scoring: {scoring.text}"
        return
    assert scoring.status_code == 200, f"persona={case['case_id']} scoring: {scoring.status_code} {scoring.text}"
    data = scoring.json()
    assert data["score"] == int(case["score"])
    _assert_close(case, "pd", data["pd"], _num(case, "pd"), PD_TOLERANCE)
    _assert_close(case, "equilibriumRate", data["equilibriumRate"], _num(case, "rate"), RATE_TOLERANCE)
    _assert_close(case, "pGap", data["cashGap"]["pGap"], _num(case, "p_gap"), PGAP_TOLERANCE)


@pytest.mark.e2e
def test_readiness_reports_all_dependencies(live_client: httpx.Client) -> None:
    response = live_client.get("/ready")
    assert response.status_code == 200, response.text
    assert response.json()["checks"] == {"db": "ok", "ml": "ok"}
