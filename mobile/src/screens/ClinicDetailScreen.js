import { useEffect, useState } from "react";
import { View, Text, FlatList, TouchableOpacity, StyleSheet } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { api } from "../api";
import { colors, radius, shadow } from "../theme";
import { SkeletonCard } from "../components/Skeleton";
import Badge from "../components/Badge";

export default function ClinicDetailScreen({ route, navigation }) {
  const { clinicId, name } = route.params;
  const [data, setData] = useState(null);

  useEffect(() => {
    navigation.setOptions({ title: name });
    api.clinicDetail(clinicId).then(setData);
  }, [clinicId]);

  return (
    <View style={styles.screen}>
      {data && (
        <View style={styles.info}>
          <View style={styles.infoRow}>
            <Ionicons name="location" size={15} color={colors.muted} />
            <Text style={styles.address}>{data.clinic.Address}</Text>
          </View>
          <View style={{ flexDirection: "row", gap: 8, marginTop: 10 }}>
            <Badge label={`⭐ ${data.clinic.Rating}`} tone="gold" />
            <Badge label={`${data.categories.length} направлений`} tone="purple" />
          </View>
          <Text style={styles.desc}>{data.clinic.Description}</Text>
        </View>
      )}
      <FlatList
        data={data ? data.categories : Array.from({ length: 6 })}
        keyExtractor={(c, i) => (c ? c.Category : String(i))}
        contentContainerStyle={{ padding: 16, gap: 12 }}
        renderItem={({ item }) =>
          !data ? (
            <SkeletonCard />
          ) : (
            <TouchableOpacity
              activeOpacity={0.8}
              style={styles.card}
              onPress={() =>
                navigation.navigate("Category", {
                  clinicId,
                  clinicName: name,
                  category: item.Category,
                })
              }
            >
              <View style={styles.emojiWrap}>
                <Text style={styles.emoji}>{item.Emoji}</Text>
              </View>
              <View style={{ flex: 1 }}>
                <Text style={styles.cardTitle}>{item.Category}</Text>
                <Text style={styles.cardMeta}>
                  {item.Count} специалиста · от {item.MinPrice.toLocaleString("ru-RU")} ₸ · ⭐ {item.MaxRating}
                </Text>
              </View>
              <Ionicons name="chevron-forward" size={18} color={colors.faint} />
            </TouchableOpacity>
          )
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg },
  info: { padding: 20, backgroundColor: colors.card, borderBottomWidth: 1, borderBottomColor: colors.border },
  infoRow: { flexDirection: "row", alignItems: "center", gap: 6 },
  address: { fontSize: 14, color: colors.ink, fontWeight: "600" },
  desc: { fontSize: 12.5, color: colors.muted, marginTop: 10 },
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
  emojiWrap: {
    width: 44,
    height: 44,
    borderRadius: radius.md,
    backgroundColor: "#f1edff",
    alignItems: "center",
    justifyContent: "center",
  },
  emoji: { fontSize: 22 },
  cardTitle: { fontSize: 15, fontWeight: "700", color: colors.ink },
  cardMeta: { fontSize: 12, color: colors.muted, marginTop: 2 },
});
