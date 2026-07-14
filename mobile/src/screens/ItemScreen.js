import { useEffect, useMemo, useState } from "react";
import { View, Text, SectionList, TouchableOpacity, StyleSheet } from "react-native";
import { LinearGradient } from "expo-linear-gradient";
import { Ionicons } from "@expo/vector-icons";
import { api } from "../api";
import { colors, gradients, radius, shadow } from "../theme";
import { useAuth } from "../AuthContext";
import EmptyState from "../components/EmptyState";

function dayLabel(iso) {
  const d = new Date(iso);
  const today = new Date();
  const tomorrow = new Date();
  tomorrow.setDate(today.getDate() + 1);
  const sameDay = (a, b) => a.toDateString() === b.toDateString();
  if (sameDay(d, today)) return "Сегодня";
  if (sameDay(d, tomorrow)) return "Завтра";
  return d.toLocaleDateString("ru-RU", { weekday: "long", day: "2-digit", month: "long" });
}

function timeLabel(iso) {
  return new Date(iso).toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" });
}

export default function ItemScreen({ route, navigation }) {
  const { itemId, clinicName } = route.params;
  const { user } = useAuth();
  const [data, setData] = useState(null);

  useEffect(() => {
    api.itemDetail(itemId).then(setData);
  }, [itemId]);

  const sections = useMemo(() => {
    if (!data) return [];
    const byDay = {};
    for (const slot of data.slots) {
      const label = dayLabel(slot.When);
      if (!byDay[label]) byDay[label] = [];
      byDay[label].push(slot);
    }
    return Object.entries(byDay).map(([title, slots]) => ({ title, data: [slots] }));
  }, [data]);

  if (!data) {
    return <View style={styles.screen} />;
  }

  const { item } = data;

  return (
    <SectionList
      style={styles.screen}
      sections={sections}
      keyExtractor={(_, i) => String(i)}
      stickySectionHeadersEnabled={false}
      ListHeaderComponent={
        <LinearGradient colors={gradients.brand} style={styles.hero}>
          <View style={styles.avatar}>
            <Text style={styles.avatarText}>{item.Name.split(" ").map((w) => w[0]).join("")}</Text>
          </View>
          <Text style={styles.name}>{item.Name}</Text>
          <Text style={styles.clinicName}>{clinicName}</Text>
          <View style={styles.heroMetaRow}>
            <View style={styles.heroMeta}>
              <Ionicons name="star" size={13} color="#ffd76a" />
              <Text style={styles.heroMetaText}>{item.Rating}</Text>
            </View>
            <View style={styles.heroMeta}>
              <Ionicons name="time-outline" size={13} color="#fff" />
              <Text style={styles.heroMetaText}>{item.Duration}</Text>
            </View>
          </View>
          <Text style={styles.price}>{item.Price.toLocaleString("ru-RU")} ₸</Text>
          <Text style={styles.desc}>{item.Description}</Text>
        </LinearGradient>
      }
      renderSectionHeader={({ section }) => (
        <Text style={styles.sectionTitle}>{section.title}</Text>
      )}
      renderItem={({ item: slots }) => (
        <View style={styles.slotGrid}>
          {slots.map((slot) => (
            <TouchableOpacity
              key={slot.ID}
              style={styles.slot}
              onPress={() => {
                if (!user) {
                  navigation.navigate("Login");
                  return;
                }
                navigation.navigate("Booking", {
                  slotId: slot.ID,
                  item,
                  clinicName,
                  slotWhen: slot.When,
                });
              }}
            >
              <Text style={styles.slotText}>{timeLabel(slot.When)}</Text>
            </TouchableOpacity>
          ))}
        </View>
      )}
      ListEmptyComponent={
        <EmptyState icon="📅" title="Нет свободного времени" subtitle="Загляните позже или выберите другого специалиста" />
      }
      contentContainerStyle={{ paddingBottom: 32 }}
    />
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg },
  hero: {
    padding: 24,
    paddingTop: 28,
    alignItems: "center",
    borderBottomLeftRadius: radius.xl,
    borderBottomRightRadius: radius.xl,
    marginBottom: 8,
  },
  avatar: {
    width: 64,
    height: 64,
    borderRadius: 32,
    backgroundColor: "rgba(255,255,255,0.25)",
    alignItems: "center",
    justifyContent: "center",
    marginBottom: 10,
  },
  avatarText: { color: "#fff", fontWeight: "800", fontSize: 20 },
  name: { color: "#fff", fontSize: 20, fontWeight: "800" },
  clinicName: { color: "rgba(255,255,255,0.85)", fontSize: 13, marginTop: 2 },
  heroMetaRow: { flexDirection: "row", gap: 16, marginTop: 12 },
  heroMeta: { flexDirection: "row", alignItems: "center", gap: 4 },
  heroMetaText: { color: "#fff", fontSize: 13, fontWeight: "600" },
  price: { color: "#fff", fontSize: 22, fontWeight: "800", marginTop: 14 },
  desc: { color: "rgba(255,255,255,0.85)", fontSize: 12.5, marginTop: 10, textAlign: "center" },
  sectionTitle: { fontSize: 14, fontWeight: "700", color: colors.ink, paddingHorizontal: 16, paddingTop: 12, paddingBottom: 8 },
  slotGrid: { flexDirection: "row", flexWrap: "wrap", gap: 10, paddingHorizontal: 16 },
  slot: {
    minWidth: 78,
    backgroundColor: colors.card,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: radius.sm,
    padding: 12,
    alignItems: "center",
    ...shadow.soft,
  },
  slotText: { fontSize: 13, fontWeight: "700", color: colors.ink },
});
