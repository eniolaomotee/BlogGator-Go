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
        {children}
      </body>
    </html>
  )
}
