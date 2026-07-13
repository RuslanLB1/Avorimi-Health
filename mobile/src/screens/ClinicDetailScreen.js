import { useEffect, useState } from "react";
import { View, Text, FlatList, TouchableOpacity, StyleSheet, ActivityIndicator } from "react-native";
import { api } from "../api";
import { colors } from "../theme";

export default function ClinicDetailScreen({ route, navigation }) {
  const { clinicId, name } = route.params;
  const [data, setData] = useState(null);

  useEffect(() => {
    navigation.setOptions({ title: name });
    api.clinicDetail(clinicId).then(setData);
  }, [clinicId]);

  if (!data) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={colors.purple} size="large" />
      </View>
    );
  }

  return (
    <View style={styles.screen}>
      <View style={styles.info}>
        <Text style={styles.address}>{data.clinic.Address}</Text>
        <Text style={styles.meta}>⭐ {data.clinic.Rating} · {data.clinic.Description}</Text>
      </View>
      <FlatList
        data={data.categories}
        keyExtractor={(c) => c.Category}
        contentContainerStyle={{ padding: 16, gap: 12 }}
        renderItem={({ item }) => (
          <TouchableOpacity
            style={styles.card}
            onPress={() =>
              navigation.navigate("Category", {
                clinicId,
                clinicName: name,
                category: item.Category,
              })
            }
          >
            <Text style={styles.emoji}>{item.Emoji}</Text>
            <View style={{ flex: 1 }}>
              <Text style={styles.cardTitle}>{item.Category}</Text>
              <Text style={styles.cardMeta}>
                {item.Count} специалиста · от {item.MinPrice.toLocaleString("ru-RU")} ₸ · ⭐ {item.MaxRating}
              </Text>
            </View>
          </TouchableOpacity>
        )}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg },
  center: { flex: 1, alignItems: "center", justifyContent: "center", backgroundColor: colors.bg },
  info: { padding: 16, backgroundColor: colors.card, borderBottomWidth: 1, borderBottomColor: colors.border },
  address: { fontSize: 14, color: colors.ink, fontWeight: "600" },
  meta: { fontSize: 12.5, color: colors.muted, marginTop: 4 },
  card: {
    flexDirection: "row",
    gap: 12,
    backgroundColor: colors.card,
    borderRadius: 16,
    padding: 16,
    alignItems: "center",
    borderWidth: 1,
    borderColor: colors.border,
  },
  emoji: { fontSize: 26 },
  cardTitle: { fontSize: 15, fontWeight: "700", color: colors.ink },
  cardMeta: { fontSize: 12, color: colors.muted, marginTop: 2 },
});
