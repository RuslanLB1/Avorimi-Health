import { useEffect, useState } from "react";
import { View, Text, FlatList, TouchableOpacity, StyleSheet, ActivityIndicator } from "react-native";
import { api } from "../api";
import { colors } from "../theme";

export default function CategoryScreen({ route, navigation }) {
  const { clinicId, clinicName, category } = route.params;
  const [items, setItems] = useState(null);

  useEffect(() => {
    navigation.setOptions({ title: category });
    api.clinicItems(clinicId, category).then(setItems);
  }, [clinicId, category]);

  if (!items) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={colors.purple} size="large" />
      </View>
    );
  }

  return (
    <FlatList
      style={styles.screen}
      data={items}
      keyExtractor={(it) => String(it.ID)}
      contentContainerStyle={{ padding: 16, gap: 12 }}
      renderItem={({ item }) => (
        <TouchableOpacity
          style={styles.card}
          onPress={() => navigation.navigate("Item", { itemId: item.ID, clinicName })}
        >
          <Text style={styles.emoji}>{item.Emoji}</Text>
          <View style={{ flex: 1 }}>
            <Text style={styles.cardTitle}>{item.Name}</Text>
            <Text style={styles.cardMeta}>
              ⭐ {item.Rating} · {item.Duration} · {item.Price.toLocaleString("ru-RU")} ₸
            </Text>
          </View>
        </TouchableOpacity>
      )}
    />
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg },
  center: { flex: 1, alignItems: "center", justifyContent: "center", backgroundColor: colors.bg },
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
