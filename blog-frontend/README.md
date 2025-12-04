# BlogGator Web Frontend

Modern, fast, and beautiful RSS feed aggregator frontend built with **Next.js 14**, **React 18**, and **Tailwind CSS**.

## Features

✨ **Modern UI**
- Beautiful gradient design with light/dark mode support
- Smooth animations and transitions
- Fully responsive (mobile, tablet, desktop)
- Icon library integration (Lucide React)

⚡ **High Performance**
- Next.js App Router for fast navigation
- Server-side rendering & static optimization
- Optimized images and assets
- Fast API integration with Axios

🔐 **Secure & Private**
- JWT token-based authentication
- Secure localStorage management
- CORS-enabled backend integration

## Quick Start

### Prerequisites

- Node.js 16+ (with npm or yarn)
- Running BlogGator Go backend on http://localhost:8080

### Installation

```bash
# Navigate to the frontend directory
cd web-next

# Install dependencies
npm install

# (Optional) Create .env.local with your backend URL
# By default, it connects to http://localhost:8080
```

### Development

```bash
# Start the development server
npm run dev

# Open http://localhost:3000 in your browser
```

The dev server supports hot reload — edit files and see changes instantly.

### Production Build

```bash
# Build the project
npm run build

# Start the production server
npm start

# The app will run on http://localhost:3000
```

## Project Structure

```
web-next/
├── app/
│   ├── page.tsx           # Home page
│   ├── auth/
│   │   └── page.tsx       # Login & Register page
│   ├── dashboard/
│   │   └── page.tsx       # Main dashboard (feeds & posts)
│   ├── layout.tsx         # Root layout
│   └── globals.css        # Global styles & Tailwind
├── package.json           # Dependencies
├── tailwind.config.ts     # Tailwind CSS config
├── tsconfig.json          # TypeScript config
├── next.config.js         # Next.js config
└── .env.local             # Environment variables (create if needed)
```

## Configuration

### Environment Variables

Create `.env.local` in the root of `web-next/`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

Replace with your actual backend URL if deployed elsewhere.

## Features Walkthrough

### 1. **Home Page** (`/`)
- Landing page with feature overview
- Call-to-action buttons
- Automatic redirect to dashboard if already logged in

### 2. **Authentication** (`/auth`)
- Unified login/register form
- Toggle between modes
- Error handling and loading states
- JWT token storage in localStorage

### 3. **Dashboard** (`/dashboard`)

#### Posts Tab
- View all posts from followed feeds
- Click posts to open external links
- Shows feed name, title, description, and publish date
- Sortable and filterable

#### Feeds Tab
- Add new RSS feeds with URL validation
- List of all followed feeds
- Unfollow feeds with one click
- Shows feed creation date

## Styling

The project uses **Tailwind CSS** with custom utilities:

- **`btn-primary`** — Blue gradient button (primary action)
- **`btn-secondary`** — Gray button (secondary action)
- **`input-field`** — Styled input with focus ring
- **`card`** — White/dark card with shadow and hover effects
- **Custom animations** — `fadeIn`, `slideIn`

All colors support light/dark mode via Tailwind's `dark:` prefix.

## API Integration

The frontend communicates with the BlogGator Go backend:

### Endpoints Used

- `POST /api/register` — Create new user
- `POST /api/login` — Authenticate user
- `GET /api/me` — Get current user info
- `GET /api/feeds` — Fetch user's feeds
- `GET /api/posts` — Fetch posts (with filters)
- `POST /api/feeds` — Add new feed
- `DELETE /api/feeds/{feedID}/unfollow` — Unfollow a feed

All requests include the JWT token in the `Authorization` header.

## Performance Tips

1. **Lazy Loading** — Images and components load on-demand
2. **Caching** — Next.js caches static pages automatically
3. **Minification** — Production builds are optimized
4. **Code Splitting** — Each route loads only necessary code

## Deployment

### Vercel (Recommended)

```bash
# Push your code to GitHub, then connect to Vercel
# https://vercel.com/new

# Set environment variable in Vercel Dashboard:
NEXT_PUBLIC_API_URL=<your-backend-url>
```

### Docker

```bash
# Build Docker image
docker build -t bloggator-web .

# Run container
docker run -p 3000:3000 bloggator-web
```

### Manual Hosting

```bash
npm run build
npm start
```

Deploy the `.next` folder to your hosting provider.

## Troubleshooting

### "Failed to fetch" errors
- Ensure BlogGator Go backend is running on http://localhost:8080
- Check `NEXT_PUBLIC_API_URL` in `.env.local`
- Verify backend CORS is enabled (check `api/middleware.go`)

### Blank pages after login
- Check browser console for errors
- Verify token is saved in localStorage
- Ensure backend returns valid JWT

### Styles not loading
- Run `npm install` to get Tailwind CSS
- Rebuild with `npm run build`
- Clear `.next` folder: `rm -rf .next && npm run dev`

## Technologies

- **Framework**: Next.js 14
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **Icons**: Lucide React
- **HTTP Client**: Axios
- **Package Manager**: npm

