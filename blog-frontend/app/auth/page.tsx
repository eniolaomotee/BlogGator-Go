'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Mail, Lock, ArrowRight, Loader } from 'lucide-react'
import axios from 'axios'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export default function LoginPage() {
  const router = useRouter()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [isRegisterMode, setIsRegisterMode] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const endpoint = isRegisterMode ? '/api/register' : '/api/login'
      const response = await axios.post(`${API_BASE}${endpoint}`, {
        username,
        password,
      })

      if (response.data.token) {
        localStorage.setItem('bg_token', response.data.token)
        localStorage.setItem('bg_user', JSON.stringify(response.data))
        router.push('/dashboard')
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Authentication failed. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 via-purple-50 to-pink-50 dark:from-slate-950 dark:via-purple-950 dark:to-slate-950 px-4">
      <div className="w-full max-w-md animate-fadeIn">
        {/* Logo/Header */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl mb-4 shadow-lg">
            <span className="text-2xl font-bold text-white">BG</span>
          </div>
          <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent mb-2">
            BlogGator
          </h1>
          <p className="text-slate-600 dark:text-slate-400">
            Your personalized feed aggregator
          </p>
        </div>

        {/* Card */}
        <div className="card p-8 shadow-xl animate-slideIn">
          <h2 className="text-2xl font-bold mb-6">
            {isRegisterMode ? 'Create Account' : 'Welcome Back'}
          </h2>

          {error && (
            <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg flex items-start gap-3">
              <div className="text-red-600 dark:text-red-400 font-semibold text-sm">
                {error}
              </div>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Username */}
            <div className="relative">
              <Mail className="absolute left-3 top-3 w-5 h-5 text-slate-400" />
              <input
                type="text"
                placeholder="Username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="input-field pl-10"
                required
              />
            </div>

            {/* Password */}
            <div className="relative">
              <Lock className="absolute left-3 top-3 w-5 h-5 text-slate-400" />
              <input
                type="password"
                placeholder="Password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="input-field pl-10"
                required
              />
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              disabled={loading}
              className="btn-primary w-full flex items-center justify-center gap-2 mt-6"
            >
              {loading ? (
                <>
                  <Loader className="w-4 h-4 animate-spin" />
                  {isRegisterMode ? 'Creating...' : 'Signing in...'}
                </>
              ) : (
                <>
                  {isRegisterMode ? 'Create Account' : 'Sign In'}
                  <ArrowRight className="w-4 h-4" />
                </>
              )}
            </button>
          </form>

          {/* Toggle Mode */}
          <div className="mt-6 text-center text-sm text-slate-600 dark:text-slate-400">
            {isRegisterMode ? "Already have an account? " : "Don't have an account? "}
            <button
              onClick={() => setIsRegisterMode(!isRegisterMode)}
              className="text-blue-600 dark:text-blue-400 font-semibold hover:underline"
            >
              {isRegisterMode ? 'Sign In' : 'Register'}
            </button>
          </div>
        </div>

        {/* Footer */}
        <p className="text-center text-xs text-slate-500 dark:text-slate-500 mt-8">
          🔒 Your data is secure and private
        </p>
      </div>
    </div>
  )
}
