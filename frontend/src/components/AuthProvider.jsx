import { createContext, useContext, useState, useEffect } from 'react';
import { authenticate, getMe, setToken, getToken } from '../api/client';
import DevLogin from './DevLogin';

const AuthContext = createContext(null);

// Dev trie: сохраняем роль в sessionStorage чтобы пережить перезагрузки
const DEV_ROLE_KEY = 'dev_role';

// Launch token sources (in priority order):
// 1. Cached JWT + dev_role in sessionStorage (dev mode — survives navigation)
// 2. MAX Mini App: window.WebApp?.initDataUnsafe?.start_param
// 3. Dev mode: ?launch_token=... query param
// 4. Dev mode: ?role=resident|employ|stud_council
function getLaunchToken() {
  // Cached dev role from previous auth
  const cachedRole = sessionStorage.getItem(DEV_ROLE_KEY);
  if (cachedRole && getToken()) {
    return { type: 'cached', role: cachedRole };
  }

  // MAX Bridge
  try {
    const startParam = window.WebApp?.initDataUnsafe?.start_param;
    if (startParam) return { type: 'token', value: startParam };
  } catch (_) { /* not in MAX WebView */ }

  // Dev query params
  const urlParams = new URLSearchParams(window.location.search);
  const launch = urlParams.get('launch_token');
  if (launch) return { type: 'token', value: launch };

  const role = urlParams.get('role');
  if (role === 'resident' || role === 'employ' || role === 'stud_council') {
    return { type: 'dev_role', value: role };
  }

  return null;
}

// Dev-only: call internal endpoint to get a JWT directly (no launch token, 30d TTL)
async function devAuth(role) {
  const res = await fetch(`/api/internal/dev/auth?role=${encodeURIComponent(role)}`);
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `dev auth failed: ${res.status}`);
  }
  return res.json();
}

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showLogin, setShowLogin] = useState(false);

  useEffect(() => {
    const launchData = getLaunchToken();

    if (!launchData) {
      // No token at all → show dev login screen (hits /api/internal/dev/users)
      setShowLogin(true);
      setLoading(false);
      return;
    }

    // Cached dev session — JWT already in storage, reconstruct user without API call
    if (launchData.type === 'cached') {
      setUser({
        user_id: 'dev',
        person_type: launchData.role === 'resident' ? 'student' : 'employee',
        role_override: launchData.role,
      });
      setLoading(false);
      return;
    }

    if (launchData.type === 'dev_role') {
      // Dev mode — get a real JWT from the dev endpoint
      devAuth(launchData.value)
        .then((data) => {
          setToken(data.token);
          sessionStorage.setItem(DEV_ROLE_KEY, launchData.value);
          // Map the dev endpoint response: user fields + role_override from ?role=
          setUser({
            user_id: data.user.user_id,
            person_type: data.user.person_type,
            role_override: launchData.value,
          });
          setLoading(false);
        })
        .catch((err) => {
          console.error('Dev auth error:', err);
          setError(err.message || 'Dev auth failed. Is MOCK_MODE=true and backend running?');
          setLoading(false);
        });
      return;
    }

    // Real auth — exchange launch token for JWT
    authenticate(launchData.value)
      .then((profile) => {
        setUser(profile);
        setLoading(false);
      })
      .catch((err) => {
        console.error('Auth error:', err);
        setError(err.message || 'Ошибка авторизации');
        setLoading(false);
      });
  }, []);

  // Dev login callback — called when user picks a persona from DevLogin
  function handleDevLogin(userData) {
    setUser(userData);
    setShowLogin(false);
    setLoading(false);
  }

  if (loading) {
    return <div style={{ color: '#fff', padding: '2em', textAlign: 'center' }}>Загрузка...</div>;
  }

  if (showLogin) {
    return <DevLogin onAuth={handleDevLogin} />;
  }

  if (error) {
    return (
      <div style={{ color: '#dc3545', padding: '2em', textAlign: 'center' }}>
        {error}
      </div>
    );
  }

  return (
    <AuthContext.Provider value={user}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be inside AuthProvider');
  return ctx;
}

// Map person_type to Router role
export function useRole() {
  const user = useAuth();
  // Dev auth sets role_override from the ?role= query param
  if (user.role_override) return user.role_override;

  switch (user.person_type) {
    case 'student':
      return 'resident';
    case 'employee':
      if (user.employee_id) {
        // TODO: determine if stud_council via API
        return 'employ';
      }
      return 'employ';
    default:
      return 'resident';
  }
}
