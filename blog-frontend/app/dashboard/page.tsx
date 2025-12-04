'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { LogOut, Plus, Loader } from 'lucide-react'
import axios from 'axios'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

interface Feed {
  id: string
  name: string
  created_at: string
}

interface Post {
  id: string
  title: string
  url: string
  description?: string
  feed_name: string
  published_at: string
}

interface User {
  username: string
  user_id: string
}

export default function DashboardPage() {
  const router = useRouter()
  const [user, setUser] = useState<User | null>(null)
  const [feeds, setFeeds] = useState<Feed[]>([])
  const [posts, setPosts] = useState<Post[]>([])
  const [loading, setLoading] = useState(true)
  const [feedName, setFeedName] = useState('')
  const [feedUrl, setFeedUrl] = useState('')
  const [activeTab, setActiveTab] = useState<'posts' | 'feeds'>('posts')

  useEffect(() => {
    const token = localStorage.getItem('bg_token')
    const userData = localStorage.getItem('bg_user')

    if (!token) {
      router.push('/auth')
      return
    }

    if (userData) {
      setUser(JSON.parse(userData))
    }

    fetchData(token)
  }, [router])

  const fetchData = async (token: string) => {
    try {
      setLoading(true)
      const headers = { Authorization: `Bearer ${token}` }

      const [feedsRes, postsRes] = await Promise.all([
        axios.get(`${API_BASE}/api/feeds`, { headers }),
        axios.get(`${API_BASE}/api/posts?limit=50`, { headers }),
      ])

      setFeeds(feedsRes.data || [])
      setPosts(postsRes.data || [])
    } catch (err) {
      console.error('Failed to fetch data:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleAddFeed = async (e: React.FormEvent) => {
    e.preventDefault()
    const token = localStorage.getItem('bg_token')
    if (!token) return

    try {
      await axios.post(
        `${API_BASE}/api/feeds`,
        { name: feedName, url: feedUrl },
        { headers: { Authorization: `Bearer ${token}` } }
      )
      setFeedName('')
      setFeedUrl('')
      fetchData(token)
    } catch (err) {
      console.error('Failed to add feed:', err)
    }
  }

  const handleUnfollow = async (feedId: string) => {
    const token = localStorage.getItem('bg_token')
    if (!token) return

    try {
      await axios.delete(`${API_BASE}/api/feeds/${feedId}/unfollow`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      fetchData(token)
    } catch (err) {
      console.error('Failed to unfollow:', err)
    }
  }

  const handleLogout = () => {
    localStorage.removeItem('bg_token')
    localStorage.removeItem('bg_user')
    router.push('/auth')
  }

  if (loading && !user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="flex flex-col items-center gap-4">
          <Loader className="w-8 h-8 animate-spin text-blue-500" />
          <p className="text-slate-600 dark:text-slate-400">Loading your feeds...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-950 dark:to-slate-900">
      {/* Header */}
      <header className="sticky top-0 z-50 bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
              <span className="text-white font-bold">BG</span>
            </div>
            <div>
              <h1 className="font-bold text-lg">BlogGator</h1>
              <p className="text-xs text-slate-500 dark:text-slate-400">
                Logged in as <strong>{user?.username}</strong>
              </p>
            </div>
          </div>

          <button
            onClick={handleLogout}
            className="flex items-center gap-2 px-4 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded-lg transition-colors"
          >
            <LogOut className="w-4 h-4" />
            Logout
          </button>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-8">
        {/* Tabs */}
        <div className="flex gap-4 mb-8 border-b border-slate-200 dark:border-slate-700">
          <button
            onClick={() => setActiveTab('posts')}
            className={`px-4 py-3 font-medium border-b-2 transition-all ${
              activeTab === 'posts'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'
            }`}
          >
            Posts
          </button>
          <button
            onClick={() => setActiveTab('feeds')}
            className={`px-4 py-3 font-medium border-b-2 transition-all ${
              activeTab === 'feeds'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'
            }`}
          >
            Feeds ({feeds.length})
          </button>
        </div>

        {/* Posts Tab */}
        {activeTab === 'posts' && (
          <div className="space-y-4 animate-fadeIn">
            {posts.length === 0 ? (
              <div className="card p-12 text-center">
                <p className="text-slate-500 dark:text-slate-400">
                  No posts yet. Add some feeds to get started!
                </p>
              </div>
            ) : (
              posts.map((post) => (
                <a
                  key={post.id}
                  href={post.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="card p-6 hover:shadow-lg cursor-pointer group animate-slideIn"
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1">
                      <h3 className="font-bold text-lg group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors mb-2">
                        {post.title}
                      </h3>
                      {post.description && (
                        <p className="text-slate-600 dark:text-slate-400 text-sm mb-3 line-clamp-2">
                          {post.description}
                        </p>
                      )}
                      <div className="flex items-center gap-3 text-xs text-slate-500 dark:text-slate-500">
                        <span className="font-semibold text-blue-600 dark:text-blue-400">
                          {post.feed_name}
                        </span>
                        <span>•</span>
                        <span>
                          {new Date(post.published_at).toLocaleDateString()}
                        </span>
                      </div>
                    </div>
                    <div className="text-blue-500 opacity-0 group-hover:opacity-100 transition-opacity">
                      →
                    </div>
                  </div>
                </a>
              ))
            )}
          </div>
        )}

        {/* Feeds Tab */}
        {activeTab === 'feeds' && (
          <div className="space-y-6 animate-fadeIn">
            {/* Add Feed Form */}
            <div className="card p-6">
              <h3 className="font-bold text-lg mb-4 flex items-center gap-2">
                <Plus className="w-5 h-5" />
                Add New Feed
              </h3>
              <form onSubmit={handleAddFeed} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-2">
                    Feed Name
                  </label>
                  <input
                    type="text"
                    value={feedName}
                    onChange={(e) => setFeedName(e.target.value)}
                    placeholder="e.g., Tech News"
                    className="input-field"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-2">
                    Feed URL (RSS)
                  </label>
                  <input
                    type="url"
                    value={feedUrl}
                    onChange={(e) => setFeedUrl(e.target.value)}
                    placeholder="https://example.com/feed.xml"
                    className="input-field"
                    required
                  />
                </div>
                <button type="submit" className="btn-primary w-full">
                  Add Feed
                </button>
              </form>
            </div>

            {/* Feeds List */}
            <div className="space-y-3">
              <h3 className="font-bold text-lg">Your Feeds</h3>
              {feeds.length === 0 ? (
                <div className="card p-8 text-center text-slate-500 dark:text-slate-400">
                  No feeds yet. Add one above to get started!
                </div>
              ) : (
                feeds.map((feed) => (
                  <div
                    key={feed.id}
                    className="card p-4 flex items-center justify-between hover:shadow-md animate-slideIn"
                  >
                    <div>
                      <h4 className="font-semibold">{feed.name}</h4>
                      <p className="text-xs text-slate-500 dark:text-slate-500">
                        Added {new Date(feed.created_at).toLocaleDateString()}
                      </p>
                    </div>
                    <button
                      onClick={() => handleUnfollow(feed.id)}
                      className="btn-secondary"
                    >
                      Unfollow
                    </button>
                  </div>
                ))
              )}
            </div>
          </div>
        )}
      </main>
    </div>
  )
}
