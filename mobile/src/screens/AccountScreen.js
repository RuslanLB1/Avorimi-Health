import { useState, useCallback } from "react";
import { View, Text, FlatList, StyleSheet, TouchableOpacity } from "react-native";
import { useFocusEffect } from "@react-navigation/native";
import { LinearGradient } from "expo-linear-gradient";
import { Ionicons } from "@expo/vector-icons";
import { api } from "../api";
import { colors, gradients, radius, shadow } from "../theme";
import { useAuth } from "../AuthContext";
import Card from "../components/Card";
import Badge from "../components/Badge";
import GradientButton from "../components/GradientButton";
import EmptyState from "../components/EmptyState";
import { SkeletonCard } from "../components/Skeleton";

const statusTone = { pending: "gold", paid: "teal", free_by_subscription: "purple" };
const statusLabel = {
  pending: "Ожидает оплаты",
  paid: "Оплачено",
  free_by_subscription: "По подписке",
};

export default function AccountScreen({ navigation }) {
  const { user, token, signOut } = useAuth();
  const [bookings, setBookings] = useState(null);
  const [me, setMe] = useState(null);

  useFocusEffect(
    useCallback(() => {
      if (token) {
        api.myBookings(token).then(setBookings);
        api.me(token).then(setMe);
      }
    }, [token])
  );

  if (!user) {
    return (
      <View style={styles.center}>
        <Ionicons name="person-circle-outline" size={64} color={colors.faint} />
        <Text style={styles.emptyTitle}>Войдите в аккаунт</Text>
        <Text style={styles.meta}>Чтобы видеть свои записи, анализы и подписку</Text>
        <GradientButton title="Войти" onPress={() => navigation.navigate("Login")} style={{ marginTop: 16, width: 200 }} />
      </View>
    );
  }

  return (
    <FlatList
      style={styles.screen}
      data={bookings || Array.from({ length: 3 })}
      keyExtractor={(b, i) => (b ? String(b.ID) : String(i))}
      contentContainerStyle={{ paddingBottom: 24 }}
      ListHeaderComponent={
        <>
          <LinearGradient colors={gradients.brand} style={styles.header}>
            <View style={styles.avatar}>
              <Text style={styles.avatarText}>{user.fullName[0]}</Text>
            </View>
            <Text style={styles.name}>{user.fullName}</Text>
            <Text style={styles.phone}>{user.phone}</Text>
            <TouchableOpacity style={styles.logout} onPress={signOut}>
              <Ionicons name="log-out-outline" size={14} color="#fff" />
              <Text style={styles.logoutText}>Выйти</Text>
            </TouchableOpacity>
          </LinearGradient>

          <View style={styles.body}>
            {me?.subscription ? (
              <Card style={styles.subCard}>
                <Ionicons name="ribbon" size={22} color={colors.teal} />
                <View style={{ flex: 1 }}>
                  <Text style={styles.subTitle}>{me.subscription.planName}</Text>
                  <Text style={styles.subMeta}>Осталось визитов: {me.subscription.visitsLeft}</Text>
                </View>
              </Card>
            ) : (
              <TouchableOpacity onPress={() => navigation.navigate("SubscriptionsTab")}>
                <Card style={styles.subCard}>
                  <Ionicons name="ribbon-outline" size={22} color={colors.muted} />
                  <View style={{ flex: 1 }}>
                    <Text style={styles.subTitle}>Нет активной подписки</Text>
                    <Text style={styles.subMeta}>Оформите — и визиты станут бесплатными</Text>
                  </View>
                  <Ionicons name="chevron-forward" size={18} color={colors.faint} />
                </Card>
              </TouchableOpacity>
            )}

            <View style={styles.quickRow}>
              <TouchableOpacity style={styles.quickCard} onPress={() => navigation.navigate("Results")}>
                <Ionicons name="flask-outline" size={20} color={colors.purple} />
                <Text style={styles.quickLabel}>Мои анализы</Text>
              </TouchableOpacity>
              <TouchableOpacity style={styles.quickCard} onPress={() => navigation.navigate("SubscriptionsTab")}>
                <Ionicons name="ribbon-outline" size={20} color={colors.purple} />
                <Text style={styles.quickLabel}>Подписки</Text>
              </TouchableOpacity>
            </View>

            <Text style={styles.sectionTitle}>Мои записи</Text>
          </View>
        </>
      }
      ListEmptyComponent={
        bookings && <EmptyState icon="📋" title="Пока нет записей" subtitle="Найдите специалиста и запишитесь на приём" />
      }
      renderItem={({ item: b }) =>
        !bookings ? (
          <View style={{ paddingHorizontal: 20, marginBottom: 12 }}>
            <SkeletonCard />
          </View>
        ) : (
          <View style={{ paddingHorizontal: 20, marginBottom: 12 }}>
            <Card style={styles.bookingCard}>
              <View style={{ flex: 1 }}>
                <Text style={styles.bookingTitle}>{b.item?.Name}</Text>
                <Text style={styles.meta}>{b.clinic?.Name}</Text>
                <Text style={styles.meta}>
                  {new Date(b.slot?.When).toLocaleString("ru-RU", { day: "2-digit", month: "long", hour: "2-digit", minute: "2-digit" })}
                </Text>
              </View>
              <Badge label={statusLabel[b.Status] || b.Status} tone={statusTone[b.Status] || "muted"} />
            </Card>
          </View>
        )
      }
    />
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg },
  center: { flex: 1, alignItems: "center", justifyContent: "center", backgroundColor: colors.bg, padding: 24 },
  header: {
    alignItems: "center",
    paddingTop: 58,
    paddingBottom: 28,
    borderBottomLeftRadius: radius.xl,
    borderBottomRightRadius: radius.xl,
  },
  avatar: {
    width: 64,
    height: 64,
    borderRadius: 32,
    backgroundColor: "rgba(255,255,255,0.25)",
    alignItems: "center",
    justifyContent: "center",
  },
  avatarText: { color: "#fff", fontWeight: "800", fontSize: 22 },
  name: { color: "#fff", fontSize: 18, fontWeight: "800", marginTop: 10 },
  phone: { color: "rgba(255,255,255,0.85)", fontSize: 13, marginTop: 2 },
  logout: {
    flexDirection: "row",
    alignItems: "center",
    gap: 5,
    marginTop: 14,
    paddingVertical: 6,
    paddingHorizontal: 14,
    borderRadius: radius.pill,
    backgroundColor: "rgba(255,255,255,0.2)",
  },
  logoutText: { color: "#fff", fontSize: 12.5, fontWeight: "700" },
  body: { padding: 20, gap: 14 },
  subCard: { flexDirection: "row", alignItems: "center", gap: 12 },
  subTitle: { fontSize: 14, fontWeight: "700", color: colors.ink },
  subMeta: { fontSize: 12, color: colors.muted, marginTop: 2 },
  quickRow: { flexDirection: "row", gap: 12 },
  quickCard: {
    flex: 1,
    backgroundColor: colors.card,
    borderRadius: radius.lg,
    borderWidth: 1,
    borderColor: colors.border,
    padding: 16,
    alignItems: "center",
    gap: 8,
    ...shadow.soft,
  },
  quickLabel: { fontSize: 12.5, fontWeight: "700", color: colors.ink },
  sectionTitle: { fontSize: 15, fontWeight: "800", color: colors.ink, marginTop: 4 },
  bookingCard: { flexDirection: "row", alignItems: "center", gap: 12 },
  bookingTitle: { fontSize: 14.5, fontWeight: "700", color: colors.ink },
  meta: { fontSize: 12, color: colors.muted, marginTop: 2 },
  emptyTitle: { fontSize: 16, fontWeight: "700", color: colors.ink, marginTop: 12 },
});
