import { useEffect, useState } from "react";
import { View, Text, FlatList, TouchableOpacity, StyleSheet } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { api } from "../api";
import { colors, radius, shadow } from "../theme";
import { SkeletonCard } from "../components/Skeleton";

export default function CategoryScreen({ route, navigation }) {
  const { clinicId, clinicName, category } = route.params;
  const [items, setItems] = useState(null);

  useEffect(() => {
    navigation.setOptions({ title: category });
    api.clinicItems(clinicId, category).then(setItems);
  }, [clinicId, category]);

  return (
    <FlatList
      style={styles.screen}
      data={items || Array.from({ length: 4 })}
      keyExtractor={(it, i) => (it ? String(it.ID) : String(i))}
      contentContainerStyle={{ padding: 16, gap: 12 }}
      renderItem={({ item }) =>
        !items ? (
          <SkeletonCard />
        ) : (
          <TouchableOpacity
            activeOpacity={0.8}
            style={styles.card}
            onPress={() => navigation.navigate("Item", { itemId: item.ID, clinicName })}
          >
            <View style={styles.avatar}>
              <Text style={styles.avatarText}>{item.Name.split(" ").map((w) => w[0]).join("")}</Text>
            </View>
            <View style={{ flex: 1 }}>
              <Text style={styles.cardTitle}>{item.Name}</Text>
              <View style={styles.metaRow}>
                <Ionicons name="star" size={12} color={colors.gold} />
                <Text style={styles.metaText}>{item.Rating}</Text>
                <Text style={styles.dot}>·</Text>
                <Text style={styles.metaText}>{item.Duration}</Text>
              </View>
              <Text style={styles.price}>{item.Price.toLocaleString("ru-RU")} ₸</Text>
            </View>
            <Ionicons name="chevron-forward" size={18} color={colors.faint} />
          </TouchableOpacity>
        )
      }
    />
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg },
  card: {
    flexDirection: "row",
    gap: 12,
    backgroundColor: colors.card,
    borderRadius: radius.lg,
    padding: 14,
    alignItems: "center",
    borderWidth: 1,
    borderColor: colors.border,
    ...shadow.soft,
  },
  avatar: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: colors.purple,
    alignItems: "center",
    justifyContent: "center",
  },
  avatarText: { color: "#fff", fontWeight: "800", fontSize: 14 },
  cardTitle: { fontSize: 15, fontWeight: "700", color: colors.ink },
  metaRow: { flexDirection: "row", alignItems: "center", gap: 4, marginTop: 3 },
  metaText: { fontSize: 12, color: colors.muted, fontWeight: "600" },
  dot: { color: colors.faint },
  price: { fontSize: 13, fontWeight: "800", color: colors.purpleDark, marginTop: 4 },
});
