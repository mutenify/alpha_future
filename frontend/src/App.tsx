import { FormEvent, useEffect, useState } from 'react'

type DemoUser = { login: string; fullName: string; role: string; scenarioHint?: string }
type Product = { code: string; name: string; description: string; minAmount: number; maxAmount: number; minTermMonths: number; maxTermMonths: number; baseRate: number }
type Offer = { id: string; variant: string; amount: number; termMonths: number; rate: number; monthlyPayment: number; pGap?: number }
type Application = { id: string; number: number; status: string; productCode: string; amount: number; termMonths: number; decisionReason?: string; offers?: Offer[] }

const API = '/api/v1'
const innByLogin: Record<string, string> = {
  client1:'7701234560', client2:'7812345675', client3:'5401234566', client4:'6671234569', client5:'5251234564',
  client6:'7700000503', client7:'1651234561', client8:'2311234566', client9:'6161234561', client10:'7707083893'
}

async function request<T>(path: string, token = '', init: RequestInit = {}): Promise<T> {
  const response = await fetch(API + path, { ...init, headers: { 'Content-Type':'application/json', ...(token ? {Authorization:`Bearer ${token}`} : {}), ...init.headers } })
  const body = await response.json().catch(() => ({}))
  if (!response.ok) throw new Error(body.message || body.error || `HTTP ${response.status}`)
  return body as T
}

export default function App() {
  const [users, setUsers] = useState<DemoUser[]>([]), [login, setLogin] = useState('client1')
  const [token, setToken] = useState(''), [products, setProducts] = useState<Product[]>([]), [apps, setApps] = useState<Application[]>([])
  const [inn, setInn] = useState(innByLogin.client1), [productCode, setProduct] = useState('WORKING_CAPITAL')
  const [amount, setAmount] = useState(300000), [term, setTerm] = useState(12), [purpose, setPurpose] = useState('Пополнение оборотных средств')
  const [error, setError] = useState(''), [busy, setBusy] = useState(false)

  useEffect(() => { request<DemoUser[]>('/auth/demo-users').then(setUsers).catch(e => setError(e.message)) }, [])
  useEffect(() => { if (innByLogin[login]) setInn(innByLogin[login]) }, [login])

  const refresh = async (auth = token) => {
    const [ps, applications] = await Promise.all([request<Product[]>('/products', auth), request<Application[]>('/applications', auth)])
    setProducts(ps); setApps(applications)
  }
  const signIn = async () => {
    setBusy(true); setError('')
    try { const session = await request<{accessToken:string}>('/auth/login', '', {method:'POST', body:JSON.stringify({login,password:'demo'})}); setToken(session.accessToken); await refresh(session.accessToken) }
    catch (e) { setError((e as Error).message) } finally { setBusy(false) }
  }
  const create = async (event: FormEvent) => {
    event.preventDefault(); setBusy(true); setError('')
    try { await request('/applications', token, {method:'POST', headers:{'Idempotency-Key':crypto.randomUUID()}, body:JSON.stringify({companyInn:inn,productCode,amount,termMonths:term,purpose})}); await refresh() }
    catch (e) { setError((e as Error).message) } finally { setBusy(false) }
  }
  const submit = async (id: string) => {
    setBusy(true); setError('')
    try { await request(`/applications/${id}/submit`, token, {method:'POST', headers:{'Idempotency-Key':crypto.randomUUID()}}); await refresh(); setTimeout(() => refresh().catch(()=>{}), 2500) }
    catch (e) { setError((e as Error).message) } finally { setBusy(false) }
  }

  return <main>
    <header><div><small>ДЕМО-СТЕНД</small><h1>Умный кредит для бизнеса</h1></div><span className="status">● backend через /api/v1</span></header>
    {error && <div className="error">{error}</div>}
    {!token ? <section className="panel login"><h2>Демо-вход</h2><select value={login} onChange={e=>setLogin(e.target.value)}>{users.map(u=><option key={u.login} value={u.login}>{u.fullName} · {u.role}</option>)}</select><p>{users.find(u=>u.login===login)?.scenarioHint}</p><button disabled={busy} onClick={signIn}>Войти, пароль demo</button></section> : <>
      <section><h2>Продукты</h2><div className="cards">{products.map(p=><article key={p.code}><b>{p.name}</b><p>{p.description}</p><small>{p.minAmount.toLocaleString('ru')}–{p.maxAmount.toLocaleString('ru')} ₽ · {p.minTermMonths}–{p.maxTermMonths} мес. · от {p.baseRate}%</small></article>)}</div></section>
      <section className="panel"><h2>Новая заявка</h2><form onSubmit={create}><label>ИНН<input value={inn} onChange={e=>setInn(e.target.value)}/></label><label>Продукт<select value={productCode} onChange={e=>setProduct(e.target.value)}>{products.map(p=><option key={p.code}>{p.code}</option>)}</select></label><label>Сумма<input type="number" value={amount} onChange={e=>setAmount(+e.target.value)}/></label><label>Срок<input type="number" value={term} onChange={e=>setTerm(+e.target.value)}/></label><label className="wide">Цель<input value={purpose} onChange={e=>setPurpose(e.target.value)}/></label><button disabled={busy}>Создать черновик</button></form></section>
      <section><div className="row"><h2>Мои заявки</h2><button className="secondary" onClick={()=>refresh()}>Обновить</button></div>{apps.length===0?<p>Заявок пока нет.</p>:apps.map(a=><article className="application" key={a.id}><div><b>Z-{a.number}</b><span className={`badge ${a.status}`}>{a.status}</span><p>{a.productCode} · {a.amount.toLocaleString('ru')} ₽ · {a.termMonths} мес.</p>{a.decisionReason&&<p>{a.decisionReason}</p>}</div>{a.status==='DRAFT'&&<button disabled={busy} onClick={()=>submit(a.id)}>Отправить</button>}{a.offers?.map(o=><div className="offer" key={o.id}><b>{o.variant}</b> {o.amount.toLocaleString('ru')} ₽ / {o.termMonths} мес. · {o.rate}% · платёж {o.monthlyPayment.toLocaleString('ru')} ₽</div>)}</article>)}</section>
    </>}
  </main>
}
