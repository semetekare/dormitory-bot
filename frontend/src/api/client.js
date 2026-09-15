const API_BASE = '/api/v1';

function getToken() {
  return sessionStorage.getItem('jwt');
}

function setToken(token) {
  sessionStorage.setItem('jwt', token);
}

export { setToken, getToken };

function clearToken() {
  sessionStorage.removeItem('jwt');
}

async function request(path, options = {}) {
  const token = getToken();
  const headers = { 'Content-Type': 'application/json', ...options.headers };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

  if (res.status === 401) {
    clearToken();
    throw new Error('Unauthorized');
  }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return res.json();
}

// ── Auth ──

export async function authenticate(launchToken) {
  const data = await request('/auth/miniapp', {
    method: 'POST',
    body: JSON.stringify({ launch_token: launchToken }),
  });
  setToken(data.token);
  return data.user;
}

export async function getMe() {
  return request('/auth/me');
}

// ── Rooms ──

export async function searchRooms(query) {
  return request(`/rooms/search?q=${encodeURIComponent(query)}`);
}

export async function getRoom(id) {
  return request(`/rooms/${id}`);
}

export async function getRoomResidents(roomId) {
  return request(`/rooms/${roomId}/residents`);
}

export async function getRoomItems(roomId) {
  return request(`/rooms/${roomId}/items`);
}

// ── Floors ──

export async function getFloors(dormitoryId) {
  return request(`/dormitories/${dormitoryId}/floors`);
}

export async function getRoomsByFloor(floorId) {
  return request(`/floors/${floorId}/rooms`);
}

// ── Laundry ──

export async function getMachines() {
  return request('/laundry/machines');
}

export async function getLaundrySlots() {
  return request('/laundry/slots');
}

export async function createBooking(payload) {
  return request('/laundry/bookings', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function cancelBooking(id) {
  return request(`/laundry/bookings/${id}`, { method: 'DELETE' });
}

// ── Cleaning ──

export async function getCleaningSchedule(dormitoryId, month, year) {
  const params = new URLSearchParams({ dormitory_id: dormitoryId });
  if (month) params.set('month', month);
  if (year) params.set('year', year);
  return request(`/cleaning/schedule?${params}`);
}

// ── References & Chat links ──

export async function getReferenceMaterials() {
  return request('/reference-materials/');
}

export async function getChatLinks() {
  return request('/chat-links/');
}

// ── Utility ──

export function isAuthenticated() {
  return !!getToken();
}

export function logout() {
  clearToken();
}
