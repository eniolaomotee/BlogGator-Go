'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { ArrowRight, Zap, Layers, Shield } from 'lucide-react'

export default function Home() {
  const router = useRouter()

  useEffect(() => {
    const token = localStorage.getItem('bg_token')
    if (token) {
      router.push('/dashboard')
    }
  }, [router])

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-purple-50 to-pink-50 dark:from-slate-950 dark:via-purple-950 dark:to-slate-950">
      {/* Navigation */}
      <nav className="sticky top-0 z-50 bg-white/80 dark:bg-slate-800/80 backdrop-blur border-b border-slate-200 dark:border-slate-700">
        <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
              <span className="text-white font-bold text-sm">BG</span>
            </div>
            <span className="font-bold text-lg">BlogGator</span>
          </div>
          <button
            onClick={() => router.push('/auth')}
            className="btn-primary flex items-center gap-2"
          >
            Get Started <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </nav>

      {/* Hero */}
      <section className="max-w-6xl mx-auto px-4 py-20 text-center">
        <h1 className="text-5xl md:text-6xl font-bold bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent mb-6 animate-fadeIn">
          Your All-in-One Feed Aggregator
        </h1>
        <p className="text-xl text-slate-600 dark:text-slate-400 mb-8 max-w-2xl mx-auto animate-slideIn">
          Discover, organize, and read all your favorite RSS feeds in one beautiful, fast interface. Stay updated without the clutter.
        </p>
        <button
          onClick={() => router.push('/auth')}
          className="btn-primary inline-flex items-center gap-2 text-lg px-8 py-4 animate-slideIn"
        >
          Start Reading Now <ArrowRight className="w-5 h-5" />
        </button>
      </section>

      {/* Features */}
      <section className="max-w-6xl mx-auto px-4 py-20">
        <h2 className="text-3xl font-bold text-center mb-16">
          Powerful Features
        </h2>
        <div className="grid md:grid-cols-3 gap-8">
          <div className="card p-8 text-center hover:shadow-lg transition-shadow animate-slideIn">
            <div className="w-12 h-12 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center mx-auto mb-4">
              <Zap className="w-6 h-6 text-blue-600 dark:text-blue-400" />
            </div>
            <h3 className="font-bold text-lg mb-2">Lightning Fast</h3>
            <p className="text-slate-600 dark:text-slate-400">
              Built with Next.js for instant loading and seamless navigation
            </p>
          </div>

          <div className="card p-8 text-center hover:shadow-lg transition-shadow animate-slideIn" style={{ animationDelay: '100ms' }}>
            <div className="w-12 h-12 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center mx-auto mb-4">
              <Layers className="w-6 h-6 text-purple-600 dark:text-purple-400" />
            </div>
            <h3 className="font-bold text-lg mb-2">Manage Feeds</h3>
            <p className="text-slate-600 dark:text-slate-400">
              Easily add, organize, and follow multiple RSS feeds with one click
            </p>
          </div>

          <div className="card p-8 text-center hover:shadow-lg transition-shadow animate-slideIn" style={{ animationDelay: '200ms' }}>
            <div className="w-12 h-12 bg-pink-100 dark:bg-pink-900 rounded-lg flex items-center justify-center mx-auto mb-4">
              <Shield className="w-6 h-6 text-pink-600 dark:text-pink-400" />
            </div>
            <h3 className="font-bold text-lg mb-2">Secure & Private</h3>
            <p className="text-slate-600 dark:text-slate-400">
              Your data is encrypted and never shared with third parties
            </p>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="max-w-4xl mx-auto px-4 py-20 text-center">
        <div className="card p-12 bg-gradient-to-r from-blue-500 to-purple-600 text-white">
          <h2 className="text-3xl font-bold mb-4">
            Ready to streamline your reading?
          </h2>
          <p className="text-lg mb-8 opacity-90">
            Join thousands of users who have simplified their news consumption
          </p>
          <button
            onClick={() => router.push('/auth')}
            className="bg-white text-blue-600 px-8 py-3 rounded-lg font-bold hover:bg-slate-100 transition-colors inline-flex items-center gap-2"
          >
            Get Started Free <ArrowRight className="w-5 h-5" />
          </button>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-slate-200 dark:border-slate-700 py-8 text-center text-slate-600 dark:text-slate-400">
        <p>© 2025 BlogGator.</p>
      </footer>
    </div>
  )
}
