import asyncio
import json
import os
from typing import Any, Literal

import httpx
from pydantic import BaseModel, Field, ValidationError


TIMEOUT_SECONDS = 2.5


class RecommendRequest(BaseModel):
    purpose: str = Field(min_length=1)
    amount: int = Field(gt=0)
    termMonths: int = Field(gt=0, le=360)


class RecommendResponse(BaseModel):
    productCode: str
    explanation: str
    source: Literal["LLM", "RULES"]


class ExplainRequest(BaseModel):
    outcome: str
    score: float
    pd: float
    pGap: float
    factors: Any = None
    pointA: Any = None
    pointB: Any = None


class ExplainResponse(BaseModel):
    decisionReason: str
    inspectorSummary: str
    source: Literal["LLM", "RULES"]


class _Recommendation(BaseModel):
    productCode: str = Field(min_length=1)
    explanation: str = Field(min_length=1)


class _Explanation(BaseModel):
    decisionReason: str = Field(min_length=1)
    inspectorSummary: str = Field(min_length=1)


def _rules_recommend(request: RecommendRequest) -> RecommendResponse:
    purpose = request.purpose.lower()
    if any(word in purpose for word in ("тендер", "контракт", "гарант")):
        code = "BANK_GUARANTEE"
        text = "Для участия в тендере или исполнения контракта подходит банковская гарантия."
    elif any(word in purpose for word in ("оборуд", "техник", "станок", "модерниз")):
        code = "INVESTMENT_LOAN"
        text = "Для покупки оборудования подходит инвестиционный кредит с погашением в течение срока проекта."
    elif any(word in purpose for word in ("кассов", "зарплат", "платеж", "разрыв")):
        code = "OVERDRAFT"
        text = "Для краткосрочного кассового разрыва подходит овердрафт по расчётному счёту."
    else:
        code = "WORKING_CAPITAL"
        text = "Для закупки товара, сырья и пополнения оборотных средств подходит оборотный кредит без залога."
    return RecommendResponse(productCode=code, explanation=text, source="RULES")


def _factor_text(factors: Any) -> str:
    if not factors:
        return "существенные дополнительные факторы не указаны"
    if isinstance(factors, list):
        values = factors[:3]
    else:
        values = [factors]
    labels = []
    for value in values:
        if isinstance(value, dict):
            labels.append(str(value.get("name") or value.get("code") or value.get("factor") or value))
        else:
            labels.append(str(value))
    return ", ".join(labels)


def _rules_explain(request: ExplainRequest) -> ExplainResponse:
    outcome = request.outcome.upper()
    pd_percent = request.pd * 100 if abs(request.pd) <= 1 else request.pd
    factors = _factor_text(request.factors)
    if outcome in {"APPROVED", "APPROVE", "ACCEPTED"}:
        reason = f"Заявка одобрена: расчётный риск {pd_percent:.1f}% соответствует требованиям банка."
    elif outcome in {"DECLINED", "DECLINE", "REJECTED", "REJECT"}:
        reason = f"Заявка отклонена: расчётный риск {pd_percent:.1f}% превышает допустимый уровень."
    else:
        reason = "Заявка направлена на ручную проверку: автоматическому решению недостаточно данных."
    summary = (
        f"Решение: {request.outcome}; score={request.score:g}; PD={request.pd:g}; "
        f"pGap={request.pGap:g}; маршрут {request.pointA} → {request.pointB}; факторы: {factors}."
    )
    return ExplainResponse(decisionReason=reason, inspectorSummary=summary, source="RULES")


def _json_from_text(text: str) -> dict:
    start, end = text.find("{"), text.rfind("}")
    if start < 0 or end < start:
        raise ValueError("LLM response does not contain JSON")
    value = json.loads(text[start : end + 1])
    if not isinstance(value, dict):
        raise ValueError("LLM response must be a JSON object")
    return value


async def _request_llm(prompt: str) -> dict | None:
    ollama_url = os.getenv("OLLAMA_URL", "").strip()
    api_url = os.getenv("LLM_API_URL", "").strip()
    api_key = os.getenv("LLM_API_KEY", "").strip()
    if ollama_url:
        url = f"{ollama_url.rstrip('/')}/api/generate"
        payload = {
            "model": os.getenv("OLLAMA_MODEL", "qwen2.5:3b"),
            "prompt": prompt,
            "format": "json",
            "stream": False,
        }
        headers = {}
        extract = lambda body: body["response"]
    elif api_url and api_key:
        url = api_url
        payload = {
            "model": os.getenv("LLM_MODEL", ""),
            "messages": [{"role": "user", "content": prompt}],
            "response_format": {"type": "json_object"},
        }
        headers = {"Authorization": f"Bearer {api_key}"}
        extract = lambda body: body["choices"][0]["message"]["content"]
    else:
        return None

    async with httpx.AsyncClient(timeout=TIMEOUT_SECONDS) as client:
        response = await client.post(url, json=payload, headers=headers)
        response.raise_for_status()
        return _json_from_text(extract(response.json()))


async def _llm_json(prompt: str) -> dict | None:
    try:
        return await asyncio.wait_for(_request_llm(prompt), timeout=TIMEOUT_SECONDS)
    except (TimeoutError, httpx.HTTPError, KeyError, TypeError, ValueError, json.JSONDecodeError):
        return None


async def recommend(request: RecommendRequest) -> RecommendResponse:
    prompt = (
        "Подбери банковский продукт для бизнеса. Ответь только JSON с полями productCode и explanation. "
        f"Цель: {request.purpose}; сумма: {request.amount}; срок, месяцев: {request.termMonths}."
    )
    generated = await _llm_json(prompt)
    if generated is not None:
        try:
            result = _Recommendation.model_validate(generated)
            return RecommendResponse(**result.model_dump(), source="LLM")
        except ValidationError:
            pass
    return _rules_recommend(request)


async def explain(request: ExplainRequest) -> ExplainResponse:
    prompt = (
        "Объясни кредитное решение простым языком. Ответь только JSON с полями decisionReason и "
        f"inspectorSummary. Данные скоринга: {request.model_dump_json()}."
    )
    generated = await _llm_json(prompt)
    if generated is not None:
        try:
            result = _Explanation.model_validate(generated)
            return ExplainResponse(**result.model_dump(), source="LLM")
        except ValidationError:
            pass
    return _rules_explain(request)
