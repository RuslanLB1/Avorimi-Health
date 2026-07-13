import { useEffect, useState, useCallback } from "react";
import { View, Text, FlatList, StyleSheet, ActivityIndicator, TouchableOpacity } from "react-native";
import { useFocusEffect } from "@react-navigation/native";
import { api } from "../api";
import { colors } from "../theme";
import { useAuth } from "../AuthContext";

const statusLabel = {
  pending: "Ожидает оплаты",
  paid: "Оплачено",
  free_by_subscription: "По подписке",
};

export default function AccountScreen({ navigation }) {
  const { user, token, signOut } = useAuth();
  const [bookings, setBookings] = useState(null);

  const load = useCallback(() => {
    if (token) api.myBookings(token).then(setBookings);
  }, [token]);

  useFocusEffect(load);

  if (!user) {
    return (
      <View style={styles.center}>
        <Text style={styles.meta}>Войдите, чтобы увидеть свои записи</Text>
        <TouchableOpacity style={styles.button} onPress={() => navigation.navigate("Login")}>
          <Text style={styles.buttonText}>Войти</Text>
        </TouchableOpacity>
      </View>
    );
  }

  if (!bookings) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={colors.purple} size="large" />
      </View>
    );
  }

  return (
    <View style={styles.screen}>
      <View style={styles.header}>
        <Text style={styles.name}>{user.fullName}</Text>
        <Text style={styles.meta}>{user.phone}</Text>
        <TouchableOpacity onPress={signOut}>
          <Text style={styles.link}>Выйти</Text>
        </TouchableOpacity>
      </View>
      <FlatList
        data={bookings}
        keyExtractor={(b) => String(b.ID)}
        contentContainerStyle={{ padding: 16, gap: 12 }}
        ListEmptyComponent={<Text style={styles.meta}>Пока нет записей</Text>}
        renderItem={({ item: b }) => (
          <View style={styles.card}>
            <Text style={styles.cardTitle}>{b.item?.Name}</Text>
            <Text style={styles.meta}>{b.clinic?.Name}</Text>
            <Text style={styles.meta}>
              {new Date(b.slot?.When).toLocaleString("ru-RU", { day: "2-digit", month: "long", hour: "2-digit", minute: "2-digit" })}
            </Text>
            <Text style={styles.status}>{statusLabel[b.Status] || b.Status}</Text>
          </View>
        )}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg },
  center: { flex: 1, alignItems: "center", justifyContent: "center", backgroundColor: colors.bg, gap: 16 },
  header: { padding: 20, backgroundColor: colors.card, borderBottomWidth: 1, borderBottomColor: colors.border },
  name: { fontSize: 18, fontWeight: "800", color: colors.ink },
  meta: { fontSize: 13, color: colors.muted, marginTop: 4 },
  link: { color: colors.purple, fontWeight: "700", marginTop: 10 },
  card: { backgroundColor: colors.card, borderRadius: 16, padding: 16, borderWidth: 1, borderColor: colors.border },
  cardTitle: { fontSize: 15, fontWeight: "700", color: colors.ink },
  status: { fontSize: 12, fontWeight: "700", color: colors.teal, marginTop: 8 },
  button: { backgroundColor: colors.purple, borderRadius: 14, padding: 14, paddingHorizontal: 28 },
  buttonText: { color: "#fff", fontWeight: "700" },
});
