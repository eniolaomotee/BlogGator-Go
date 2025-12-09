import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'BlogGator - Your Feed Aggregator',
  description: 'Discover and manage your favorite RSS feeds in one place',
  icons: {
    icon: [{ url: '/favicon.ico', sizes: 'any' }],
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <meta name="color-scheme" content="light dark" />
      </head>
      <body className="bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-950 dark:to-slate-900 text-slate-900 dark:text-slate-100">
        <header className="w-full border-b bg-white/60 dark:bg-black/40 backdrop-blur sticky top-0 z-30">
          <div className="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between">
            <div className="flex items-center gap-4">
              <a href="/" className="font-semibold text-lg">BlogGator</a>
            </div>
            <div>
              <a href="/auth" className="text-sm bg-slate-200 dark:bg-slate-800 px-3 py-1 rounded">Sign in</a>
            </div>
          </div>
        </header>
        <main className="max-w-6xl mx-auto px-4 py-8">{children}</main>
      </body>
    </html>
  )
}
