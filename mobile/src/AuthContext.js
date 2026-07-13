import { createContext, useContext, useEffect, useState } from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { api } from "./api";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [token, setToken] = useState(null);
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    (async () => {
      const saved = await AsyncStorage.getItem("avorimi_token");
      if (saved) {
        try {
          const me = await api.me(saved);
          setToken(saved);
          setUser(me.user);
        } catch {
          await AsyncStorage.removeItem("avorimi_token");
        }
      }
      setLoading(false);
    })();
  }, []);

  async function signIn(newToken, newUser) {
    await AsyncStorage.setItem("avorimi_token", newToken);
    setToken(newToken);
    setUser(newUser);
  }

  async function signOut() {
    await AsyncStorage.removeItem("avorimi_token");
    setToken(null);
    setUser(null);
  }

  return (
    <AuthContext.Provider value={{ token, user, loading, signIn, signOut }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
