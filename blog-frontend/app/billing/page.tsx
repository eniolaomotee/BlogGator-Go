"use client"

import React, { useEffect, useState } from "react"

type Tier = {
  id: string
  name: string
  price_cents: number
  features?: Record<string, any>
}

type Subscription = {
  id: string
  tier_id: string
  status: string
}

export default function BillingPage() {
  const [tiers, setTiers] = useState<Tier[]>([])
  const [sub, setSub] = useState<Subscription | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"
  const token = typeof window !== 'undefined' ? localStorage.getItem('bg_token') : null

  useEffect(() => {
    async function load() {
      setError(null)
      // Tiers
      try {
        const res = await fetch(`${apiUrl}/api/billing/tiers`, {
          headers: token ? { Authorization: `Bearer ${token}` } : undefined,
        })
        const text = await res.text()
        let data: any = null
        try {
          data = text ? JSON.parse(text) : null
        } catch (e) {
          // backend returned non-JSON (HTML/error); surface text for debugging
          throw new Error(`Invalid JSON response for tiers: ${text}`)
        }
        if (!res.ok) {
          const msg = (data && data.error) ? data.error : `HTTP ${res.status}`
          throw new Error(`Failed to fetch tiers: ${msg}`)
        }
        setTiers(data.tiers || data || [])
      } catch (e: any) {
        setError(e.message || String(e))
        setTiers([])
      }

      // Subscription
      try {
        const res = await fetch(`${apiUrl}/api/billing/subscription`, {
          headers: token ? { Authorization: `Bearer ${token}` } : undefined,
        })
        if (!res.ok) {
          setSub(null)
          return
        }
        const text = await res.text()
        try {
          const data = text ? JSON.parse(text) : null
          setSub(data || null)
        } catch (e) {
          // non-JSON subscription response
          setSub(null)
        }
      } catch (e) {
        setSub(null)
      }
    }

    load()
  }, [])

  async function handleCheckout(tierId: string) {
    setLoading(true)
    setError(null)
    try {
      const res = await fetch(`${apiUrl}/api/billing/checkout`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({ tier_id: tierId }),
      })
      if (!res.ok) {
        const text = await res.text()
        throw new Error(text || "Checkout failed")
      }
      const data = await res.json()
      // If backend returns a url for hosted checkout, redirect
      if (data.url) {
        window.location.href = data.url
        return
      }
      setSub(data.subscription || data)
    } catch (e: any) {
      setError(e.message || String(e))
    } finally {
      setLoading(false)
    }
  }

  async function handleCancel() {
    if (!sub) return
    setLoading(true)
    setError(null)
    try {
      const res = await fetch(`${apiUrl}/api/billing/cancel`, {
        method: "POST",
        headers: {
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
      })
      if (!res.ok) throw new Error("Cancel failed")
      setSub(null)
    } catch (e: any) {
      setError(e.message || String(e))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6 max-w-3xl mx-auto">
      <h1 className="text-3xl font-semibold">Billing</h1>
      {error && <div className="text-red-600">{error}</div>}

      <div className="space-y-4">
        <div className="border rounded p-4">
          <h2 className="font-medium">Your Subscription</h2>
          {sub ? (
            <div className="mt-2">
              <div>Tier: {sub.tier_id}</div>
              <div>Status: {sub.status}</div>
              <div className="mt-3">
                <button className="bg-red-600 text-white px-3 py-1 rounded" onClick={handleCancel} disabled={loading}>Cancel</button>
              </div>
            </div>
          ) : (
            <div className="mt-2">You have no active subscription.</div>
          )}
        </div>

        <div className="border rounded p-4">
          <h2 className="font-medium">Available Plans</h2>
          <div className="mt-3 space-y-3">
            {tiers.length === 0 && !error && <div className="text-sm text-slate-500">No tiers available.</div>}
            {tiers.map((t) => (
              <div key={t.id} className="flex items-center justify-between p-2 rounded bg-slate-50 dark:bg-slate-900">
                <div>
                  <div className="font-medium">{t.name}</div>
                  <div className="text-sm text-slate-500">{(t.price_cents / 100).toFixed(2)} USD / mo</div>
                </div>
                <div>
                  <button
                    className="bg-blue-600 text-white px-3 py-1 rounded disabled:opacity-50"
                    onClick={() => handleCheckout(t.id)}
                    disabled={loading}
                  >
                    {sub && sub.tier_id === t.id ? "Current" : "Select"}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
