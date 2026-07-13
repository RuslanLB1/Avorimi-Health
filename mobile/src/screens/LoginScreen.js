import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, StyleSheet, ActivityIndicator, Alert } from "react-native";
import { api } from "../api";
import { colors } from "../theme";
import { useAuth } from "../AuthContext";

export default function LoginScreen({ navigation }) {
  const { signIn } = useAuth();
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit() {
    setLoading(true);
    try {
      const res = await api.login({ phoneLocal: phone, password });
      await signIn(res.token, res.user);
      navigation.goBack();
    } catch (e) {
      Alert.alert("Ошибка входа", e.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <View style={styles.screen}>
      <Text style={styles.title}>Вход</Text>

      <View style={styles.phoneRow}>
        <Text style={styles.prefix}>+7</Text>
        <TextInput
          style={styles.phoneInput}
          placeholder="701 234 56 78"
          keyboardType="phone-pad"
          value={phone}
          onChangeText={setPhone}
        />
      </View>
      <TextInput
        style={styles.input}
        placeholder="Пароль"
        secureTextEntry
        value={password}
        onChangeText={setPassword}
      />

      <TouchableOpacity style={styles.button} onPress={submit} disabled={loading}>
        {loading ? <ActivityIndicator color="#fff" /> : <Text style={styles.buttonText}>Войти</Text>}
      </TouchableOpacity>

      <TouchableOpacity onPress={() => navigation.navigate("Register")}>
        <Text style={styles.link}>Нет аккаунта? Зарегистрироваться</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg, padding: 24, justifyContent: "center" },
  title: { fontSize: 24, fontWeight: "800", color: colors.ink, marginBottom: 20 },
  input: {
    backgroundColor: colors.card,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 12,
    padding: 14,
    marginBottom: 12,
    fontSize: 15,
  },
  phoneRow: { flexDirection: "row", alignItems: "center", marginBottom: 12 },
  prefix: {
    fontSize: 15,
    fontWeight: "700",
    color: colors.ink,
    backgroundColor: colors.card,
    borderWidth: 1,
    borderColor: colors.border,
    borderRightWidth: 0,
    borderTopLeftRadius: 12,
    borderBottomLeftRadius: 12,
    padding: 14,
  },
  phoneInput: {
    flex: 1,
    backgroundColor: colors.card,
    borderWidth: 1,
    borderColor: colors.border,
    borderTopRightRadius: 12,
    borderBottomRightRadius: 12,
    padding: 14,
    fontSize: 15,
  },
  button: { backgroundColor: colors.purple, borderRadius: 14, padding: 16, alignItems: "center", marginTop: 8 },
  buttonText: { color: "#fff", fontWeight: "700", fontSize: 15 },
  link: { color: colors.purple, textAlign: "center", marginTop: 18, fontWeight: "600" },
});
