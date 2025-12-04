const BASE = 'http://localhost:8080'; // empty = same origin (change to 'http://localhost:8080' if needed)

function getToken(){return localStorage.getItem('bg_token')}
function setToken(t){localStorage.setItem('bg_token', t)}
function clearToken(){localStorage.removeItem('bg_token')}

// Frontend error logging utilities
function _getLogs(){ try{ return JSON.parse(localStorage.getItem('bg_errors')||'[]') }catch(e){return[]} }
function _setLogs(l){ localStorage.setItem('bg_errors', JSON.stringify(l.slice(0,200))) }
function addLog(entry){ const l=_getLogs(); l.unshift(entry); _setLogs(l); renderErrorLog(); }

function renderErrorLog(){
  const el = document.getElementById('error-log');
  if(!el) return;
  const logs = _getLogs();
  if(logs.length===0){ el.textContent = 'No logs'; return }
  el.textContent = logs.map((it,i)=>`#${i} ${it.time}\n${it.message}\n${it.stack||''}\nContext: ${JSON.stringify(it.context||{})}\n`).join('\n----\n');
}

function clearLogs(){ localStorage.removeItem('bg_errors'); renderErrorLog(); }

function downloadLogs(){ const logs=_getLogs(); const blob=new Blob([JSON.stringify(logs,null,2)],{type:'application/json'}); const url=URL.createObjectURL(blob); const a=document.createElement('a'); a.href=url; a.download='bg_frontend_logs.json'; document.body.appendChild(a); a.click(); a.remove(); URL.revokeObjectURL(url); }

function logFrontendError(err, ctx){
  try{
    const entry = {
      time: new Date().toISOString(),
      message: err && err.message ? err.message : String(err),
      stack: err && err.stack ? err.stack : null,
      context: ctx || {}
    };
    addLog(entry);
    console.error('FrontendLog:', entry);
  }catch(e){console.error('error logging frontend error', e)}
}

async function request(path, opts={}){
  const headers = opts.headers || {};
  headers['Content-Type'] = headers['Content-Type'] || 'application/json';
  const token = getToken();
  if(token) headers['Authorization'] = `Bearer ${token}`;
  console.log('Request:', {path, method: opts.method||'GET', token: token ? 'present' : 'missing', headers});
  let res;
  try{
    res = await fetch(BASE + path, {...opts, headers});
  }catch(e){
    // Network or CORS failure
    logFrontendError(e, {phase:'fetch', method: opts.method||'GET', path, body: opts.body, token: token ? 'present' : 'missing'});
    throw e;
  }

  const text = await res.text();
  let body = null;
  try{ body = text ? JSON.parse(text) : null }catch(e){ body = text }
  if(!res.ok){
    const msg = (body && body.error) || (typeof body === 'string' && body) || res.statusText;
    const err = new Error(msg || 'Request failed');
    logFrontendError(err, {phase:'response', status: res.status, path, body});
    throw err;
  }
  return body;
}

// Auth handlers
document.getElementById('login-form').addEventListener('submit', async (e)=>{
  e.preventDefault();
  const username = document.getElementById('login-username').value.trim();
  const password = document.getElementById('login-password').value;
  try{
    const data = await request('/api/login', {method:'POST', body:JSON.stringify({username,password})});
    setToken(data.token);
    showApp();
  }catch(err){alert('Login error: '+err.message)}
})

document.getElementById('register-form').addEventListener('submit', async (e)=>{
  e.preventDefault();
  const username = document.getElementById('register-username').value.trim();
  const password = document.getElementById('register-password').value;
  try{
    const data = await request('/api/register', {method:'POST', body:JSON.stringify({username,password})});
    setToken(data.token);
    showApp();
  }catch(err){alert('Register error: '+err.message)}
})

document.getElementById('logout-btn').addEventListener('click', ()=>{
  clearToken(); location.reload();
})

// Feeds
async function fetchFeeds(){
  try{
    const feeds = await request('/api/feeds');
    renderFeeds(feeds);
  }catch(err){alert('Error fetching feeds: '+err.message)}
}

function renderFeeds(feeds){
  const list = document.getElementById('feeds-list');
  list.innerHTML = '';
  if(!feeds || feeds.length===0){ list.innerHTML = '<li class="muted">No feeds yet</li>'; return }
  feeds.forEach(f=>{
    const li = document.createElement('li');
    const left = document.createElement('div');
    left.innerHTML = `<strong>${escapeHtml(f.name)}</strong><div class="muted">${f.created_at}</div>`;
    const right = document.createElement('div');
    const unfollow = document.createElement('button');
    unfollow.textContent = 'Unfollow';
    unfollow.addEventListener('click', ()=>unfollowFeed(f.id));
    right.appendChild(unfollow);
    li.appendChild(left);
    li.appendChild(right);
    list.appendChild(li);
  })
}

document.getElementById('add-feed-form').addEventListener('submit', async (e)=>{
  e.preventDefault();
  const name = document.getElementById('feed-name').value.trim();
  const url = document.getElementById('feed-url').value.trim();
  try{
    await request('/api/feeds', {method:'POST', body:JSON.stringify({name,url})});
    document.getElementById('feed-name').value = '';
    document.getElementById('feed-url').value = '';
    fetchFeeds();
  }catch(err){alert('Add feed error: '+err.message)}
})

async function unfollowFeed(id){
  if(!confirm('Unfollow this feed?')) return;
  try{
    await request(`/api/feeds/${id}/unfollow`, {method:'DELETE'});
    fetchFeeds();
  }catch(err){alert('Unfollow error: '+err.message)}
}

// Follow via feed id (not used in list but available)
async function followFeed(id){
  try{
    await request('/api/feeds/follow', {method:'POST', body:JSON.stringify({feed_id:id})});
    fetchFeeds();
  }catch(err){alert('Follow error: '+err.message)}
}

// Posts
async function fetchPosts(){
  const limit = document.getElementById('posts-limit').value || 20;
  const sort = document.getElementById('posts-sort').value;
  const order = document.getElementById('posts-order').value;
  const q = `?limit=${encodeURIComponent(limit)}&sort=${encodeURIComponent(sort)}&order=${encodeURIComponent(order)}`;
  try{
    const posts = await request('/api/posts'+q);
    renderPosts(posts);
  }catch(err){alert('Error fetching posts: '+err.message)}
}

function renderPosts(posts){
  const list = document.getElementById('posts-list');
  list.innerHTML = '';
  if(!posts || posts.length===0){ list.innerHTML = '<li class="muted">No posts</li>'; return }
  posts.forEach(p=>{
    const li = document.createElement('li');
    const left = document.createElement('div');
    const title = `<a class="post-link" href="${escapeHtml(p.url)}" target="_blank">${escapeHtml(p.title)}</a>`;
    left.innerHTML = `${title}<div class="muted">${p.feed_name} • ${new Date(p.published_at).toLocaleString()}</div>`;
    li.appendChild(left);
    list.appendChild(li);
  })
}

document.getElementById('refresh-posts').addEventListener('click', fetchPosts);

function showApp(){
  document.getElementById('auth-section').classList.add('hidden');
  document.getElementById('app-section').classList.remove('hidden');
  fetchCurrentUser();
  fetchFeeds();
  fetchPosts();
}

async function fetchCurrentUser(){
  try{
    const me = await request('/api/me');
    document.getElementById('user-info').textContent = me.username || '';
  }catch(e){console.warn('no current user')}
}

function escapeHtml(s){ if(!s) return ''; return s.replace(/[&<>"]/g, c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c])) }

// On load
window.addEventListener('DOMContentLoaded', ()=>{
  if(getToken()) showApp();
  // Wire debug buttons
  const toggle = document.getElementById('toggle-logs');
  const clearBtn = document.getElementById('clear-logs');
  const downloadBtn = document.getElementById('download-logs');
  if(toggle){
    toggle.addEventListener('click', ()=>{
      const el=document.getElementById('error-log');
      if(!el) return;
      if(el.classList.contains('hidden')){
        el.classList.remove('hidden'); toggle.textContent='Hide Logs'; renderErrorLog();
      } else { el.classList.add('hidden'); toggle.textContent='Show Logs'; }
    })
  }
  if(clearBtn) clearBtn.addEventListener('click', ()=>{ if(confirm('Clear frontend logs?')) clearLogs(); });
  if(downloadBtn) downloadBtn.addEventListener('click', ()=>{ downloadLogs(); });
})
