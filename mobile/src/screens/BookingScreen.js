import { useState } from "react";
import { View, Text, TouchableOpacity, StyleSheet, ActivityIndicator, Alert } from "react-native";
import { api } from "../api";
import { colors } from "../theme";
import { useAuth } from "../AuthContext";

function formatSlot(iso) {
  const d = new Date(iso);
  return d.toLocaleString("ru-RU", {
    day: "2-digit",
    month: "long",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function BookingScreen({ route, navigation }) {
  const { slotId, item, clinicName, slotWhen } = route.params;
  const { token } = useAuth();
  const [loading, setLoading] = useState(false);

  async function confirmAndPay() {
    setLoading(true);
    try {
      const booking = await api.createBooking(slotId, false, token);
      await api.payBooking(booking.ID, token);
      navigation.replace("Success", { item, clinicName, slotWhen });
    } catch (e) {
      Alert.alert("Не получилось", e.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <View style={styles.screen}>
      <View style={styles.card}>
        <Text style={styles.title}>Подтверждение записи</Text>
        <Text style={styles.row}>👤 {item.Name}</Text>
        <Text style={styles.row}>🏥 {clinicName}</Text>
        <Text style={styles.row}>🗓 {formatSlot(slotWhen)}</Text>
        <Text style={styles.price}>{item.Price.toLocaleString("ru-RU")} ₸</Text>
        <Text style={styles.note}>
          Демо-оплата: реальные банковские данные не запрашиваются и не списываются.
        </Text>
      </View>
      <TouchableOpacity style={styles.button} onPress={confirmAndPay} disabled={loading}>
        {loading ? <ActivityIndicator color="#fff" /> : <Text style={styles.buttonText}>Оплатить и подтвердить</Text>}
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg, padding: 20, justifyContent: "center" },
  card: { backgroundColor: colors.card, borderRadius: 16, padding: 20, borderWidth: 1, borderColor: colors.border },
  title: { fontSize: 18, fontWeight: "800", color: colors.ink, marginBottom: 12 },
  row: { fontSize: 14, color: colors.ink, marginTop: 6 },
  price: { fontSize: 22, fontWeight: "800", color: colors.purpleDark, marginTop: 14 },
  note: { fontSize: 12, color: colors.muted, marginTop: 14 },
  button: {
    backgroundColor: colors.purple,
    borderRadius: 14,
    padding: 16,
    alignItems: "center",
    marginTop: 20,
  },
  buttonText: { color: "#fff", fontWeight: "700", fontSize: 15 },
});
