import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, StyleSheet, ActivityIndicator, Alert, ScrollView } from "react-native";
import { api } from "../api";
import { colors } from "../theme";
import { useAuth } from "../AuthContext";

export default function RegisterScreen({ navigation }) {
  const { signIn } = useAuth();
  const [fullName, setFullName] = useState("");
  const [iin, setIin] = useState("");
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit() {
    setLoading(true);
    try {
      const res = await api.register({
        fullName,
        iin,
        phoneLocal: phone,
        password,
        confirmPassword: confirm,
      });
      await signIn(res.token, res.user);
      navigation.goBack();
    } catch (e) {
      Alert.alert("Ошибка регистрации", e.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <ScrollView contentContainerStyle={styles.screen}>
      <Text style={styles.title}>Регистрация</Text>

      <TextInput style={styles.input} placeholder="ФИО" value={fullName} onChangeText={setFullName} />
      <TextInput
        style={styles.input}
        placeholder="ИИН (12 цифр)"
        keyboardType="number-pad"
        maxLength={12}
        value={iin}
        onChangeText={setIin}
      />
      <Text style={styles.note}>В клинике могут попросить удостоверение личности для сверки.</Text>

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
      <TextInput style={styles.input} placeholder="Пароль" secureTextEntry value={password} onChangeText={setPassword} />
      <TextInput
        style={styles.input}
        placeholder="Повторите пароль"
        secureTextEntry
        value={confirm}
        onChangeText={setConfirm}
      />

      <TouchableOpacity style={styles.button} onPress={submit} disabled={loading}>
        {loading ? <ActivityIndicator color="#fff" /> : <Text style={styles.buttonText}>Зарегистрироваться</Text>}
      </TouchableOpacity>

      <TouchableOpacity onPress={() => navigation.navigate("Login")}>
        <Text style={styles.link}>Уже есть аккаунт? Войти</Text>
      </TouchableOpacity>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  screen: { flexGrow: 1, backgroundColor: colors.bg, padding: 24, justifyContent: "center" },
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
  note: { fontSize: 12, color: colors.muted, marginTop: -6, marginBottom: 12 },
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
