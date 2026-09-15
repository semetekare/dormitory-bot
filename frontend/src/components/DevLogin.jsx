import { useState, useEffect } from 'react';
import { setToken } from '../api/client';
import '../styles/DevLogin.css';

// Dev-only: call /api/internal/dev/auth?phone=... to get JWT
async function devAuthByPhone(phone) {
  const res = await fetch(`/api/internal/dev/auth?phone=${encodeURIComponent(phone)}`);
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `auth failed: ${res.status}`);
  }
  return res.json();
}

function UserCard({ user, onClick }) {
  const hasDebt = user.debt && user.debt !== '—';
  return (
    <div className="user-card" onClick={() => onClick(user.phone)}>
      <div>
        <div className="name">{user.name}</div>
        {user.title && <div className="meta">{user.title}</div>}
        {user.debt && <div className="meta">Задолженность: {user.debt}</div>}
      </div>
      {hasDebt ? (
        <span className="badge debt">{user.debt}</span>
      ) : (
        <span className="badge">{user.title ? user.title.split(' ')[0] : user.role}</span>
      )}
    </div>
  );
}

export default function DevLogin({ onAuth }) {
  const [users, setUsers] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetch('/api/internal/dev/users')
      .then((r) => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return r.json();
      })
      .then((data) => {
        setUsers(data);
        setLoading(false);
      })
      .catch((err) => {
        console.error('Dev login: failed to load users:', err);
        setError('Не удалось загрузить список пользователей. Бэкенд запущен с MOCK_MODE=true?');
        setLoading(false);
      });
  }, []);

  async function handleLogin(phone) {
    try {
      setLoading(true);
      const data = await devAuthByPhone(phone);
      setToken(data.token);
      const role = data.user.role;
      sessionStorage.setItem('dev_role', role);
      onAuth({
        user_id: data.user.user_id,
        person_type: data.user.person_type,
        role_override: role,
      });
    } catch (err) {
      console.error('Dev login error:', err);
      setError(err.message || 'Ошибка входа');
      setLoading(false);
    }
  }

  if (loading) {
    return <div className="dev-login"><div style={{textAlign:'center', color:'#aaa'}}>Загрузка...</div></div>;
  }

  if (error) {
    return <div className="dev-login"><div className="error">{error}</div></div>;
  }

  return (
    <div className="dev-login">
      <h2>🔑 Dev-вход</h2>
      <div style={{fontSize:'0.85em',color:'#888',marginBottom:'1em'}}>
        Выберите тестового пользователя
      </div>

      {users?.employ?.length > 0 && (
        <>
          <div className="section-title">Сотрудники</div>
          {users.employ.map((u) => (
            <UserCard key={u.phone} user={u} onClick={handleLogin} />
          ))}
        </>
      )}

      {users?.stud_council?.length > 0 && (
        <>
          <div className="section-title">Студсовет</div>
          {users.stud_council.map((u) => (
            <UserCard key={u.phone} user={u} onClick={handleLogin} />
          ))}
        </>
      )}

      {users?.resident?.length > 0 && (
        <>
          <div className="section-title">Проживающие</div>
          {users.resident.map((u) => (
            <UserCard key={u.phone} user={u} onClick={handleLogin} />
          ))}
        </>
      )}
    </div>
  );
}