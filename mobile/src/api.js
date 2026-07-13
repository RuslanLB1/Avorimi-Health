const API_BASE = "https://avorimi-health.onrender.com/api";

async function request(path, { method = "GET", token, body } = {}) {
  const headers = { "Content-Type": "application/json" };
  if (token) headers.Authorization = `Bearer ${token}`;

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  const data = await res.json().catch(() => null);
  if (!res.ok) {
    throw new Error((data && data.error) || `err.http${res.status}`);
  }
  return data;
}

export const api = {
  register: (payload) => request("/register", { method: "POST", body: payload }),
  login: (payload) => request("/login", { method: "POST", body: payload }),
  me: (token) => request("/me", { token }),

  clinics: (lat, lng) =>
    request(lat && lng ? `/clinics?lat=${lat}&lng=${lng}` : "/clinics"),
  clinicDetail: (id) => request(`/clinics/${id}`),
  clinicItems: (id, category) =>
    request(`/clinics/${id}/items?category=${encodeURIComponent(category)}`),
  itemDetail: (id) => request(`/items/${id}`),

  plans: () => request("/plans"),
  subscribe: (planId, token) => request(`/subscribe/${planId}`, { method: "POST", token }),

  createBooking: (slotId, useSubscription, token) =>
    request("/bookings", { method: "POST", token, body: { slotId, useSubscription } }),
  payBooking: (id, token) => request(`/bookings/${id}/pay`, { method: "POST", token }),
  myBookings: (token) => request("/bookings", { token }),
};
