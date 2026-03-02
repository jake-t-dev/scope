import { useState, FormEvent } from 'react'

interface NewsArticle {
  source: {
    id: string | null;
    name: string;
  };
  author: string | null;
  title: string;
  description: string | null;
  url: string;
  urlToImage: string | null;
  publishedAt: string;
  content: string | null;
}

function App() {
  const [username, setUsername] = useState('')
  const [articles, setArticles] = useState<NewsArticle[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchNews = async (e: FormEvent) => {
    e.preventDefault()
    if (!username.trim()) return

    setLoading(true)
    setError(null)
    setArticles([])

    try {
      const response = await fetch(`http://localhost:${import.meta.env.VITE_PORT}/?username=${encodeURIComponent(username)}`)
      if (!response.ok) {
        if (response.status === 400) throw new Error('Username is required')
        if (response.status === 500) throw new Error('Server error or user not found')
        throw new Error('Failed to fetch')
      }
      const data = await response.json()
      setArticles(data || [])
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-100 py-8 px-4">
      <div className="max-w-4xl mx-auto">
        <header className="mb-8 text-center">
          <h1 className="text-3xl font-bold text-gray-900">Your Tech News Feed</h1>
          <p className="text-gray-600 mt-2">Curated based on your GitHub interests</p>
        </header>

        <form onSubmit={fetchNews} className="mb-8 max-w-md mx-auto flex gap-2">
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="Enter GitHub Username"
            className="flex-1 px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
          <button
            type="submit"
            disabled={loading}
            className="bg-blue-600 text-white px-6 py-2 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {loading ? 'Loading...' : 'Get News'}
          </button>
        </form>

        {error && <div className="text-red-500 text-center mb-6 bg-red-50 p-3 rounded-md border border-red-200">{error}</div>}
        
        {!loading && articles.length === 0 && !error && (
          <div className="text-center text-gray-500">
            Enter a GitHub username to see personalized tech news.
          </div>
        )}
        
        <div className="space-y-6">
          {articles.map((article, i) => (
            <div key={i} className="bg-white shadow rounded-lg p-6 hover:shadow-lg transition-shadow">
              <div className="flex flex-col sm:flex-row justify-between items-start gap-4">
                <div className="flex-1">
                  <h2 className="text-xl font-semibold mb-2">
                    <a href={article.url} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">
                      {article.title}
                    </a>
                  </h2>
                  <p className="text-sm text-gray-500 mb-2">
                    {article.source.name} • {new Date(article.publishedAt).toLocaleDateString()}
                  </p>
                  <p className="text-gray-700 line-clamp-3">{article.description}</p>
                </div>
                {article.urlToImage && (
                  <img 
                    src={article.urlToImage} 
                    alt={article.title} 
                    className="w-full sm:w-48 h-32 object-cover rounded bg-gray-200 flex-shrink-0"
                    onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
                  />
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default App