export const colors = {
  purple: "#7c5cff",
  purpleDark: "#6238e0",
  blue: "#2f6fed",
  blueDark: "#1f4fbf",
  teal: "#12b8a6",
  ink: "#16213a",
  muted: "#6b7488",
  faint: "#a3a9ba",
  bg: "#f6f5fc",
  card: "#ffffff",
  border: "#e7e4f5",
  danger: "#e5484d",
  success: "#12b8a6",
  gold: "#f5a623",
};

export const gradients = {
  brand: ["#7c5cff", "#2f6fed"],
  brandSoft: ["#efeaff", "#e7f0ff"],
  teal: ["#12b8a6", "#0e8f81"],
  gold: ["#f5c453", "#f0a623"],
};

export const radius = {
  sm: 10,
  md: 14,
  lg: 18,
  xl: 24,
  pill: 999,
};

export const spacing = (n) => n * 4;

export const shadow = {
  card: {
    shadowColor: "#5a3cc8",
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.1,
    shadowRadius: 20,
    elevation: 4,
  },
  soft: {
    shadowColor: "#5a3cc8",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.06,
    shadowRadius: 10,
    elevation: 2,
  },
};

export const type = {
  h1: { fontSize: 26, fontWeight: "800" },
  h2: { fontSize: 20, fontWeight: "800" },
  h3: { fontSize: 16, fontWeight: "700" },
  body: { fontSize: 14.5, fontWeight: "400" },
  small: { fontSize: 12.5, fontWeight: "500" },
};
