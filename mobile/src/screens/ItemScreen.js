import { useEffect, useState } from "react";
import { View, Text, FlatList, TouchableOpacity, StyleSheet, ActivityIndicator } from "react-native";
import { api } from "../api";
import { colors } from "../theme";
import { useAuth } from "../AuthContext";

function formatSlot(iso) {
  const d = new Date(iso);
  const day = d.toLocaleDateString("ru-RU", { day: "2-digit", month: "short" });
  const time = d.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" });
  return `${day}, ${time}`;
}

export default function ItemScreen({ route, navigation }) {
  const { itemId, clinicName } = route.params;
  const { user } = useAuth();
  const [data, setData] = useState(null);

  useEffect(() => {
    api.itemDetail(itemId).then(setData);
  }, [itemId]);

  if (!data) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={colors.purple} size="large" />
      </View>
    );
  }

  const { item, slots } = data;

  return (
    <View style={styles.screen}>
      <View style={styles.info}>
        <Text style={styles.emoji}>{item.Emoji}</Text>
        <Text style={styles.name}>{item.Name}</Text>
        <Text style={styles.meta}>{clinicName}</Text>
        <Text style={styles.meta}>⭐ {item.Rating} · {item.Duration}</Text>
        <Text style={styles.price}>{item.Price.toLocaleString("ru-RU")} ₸</Text>
        <Text style={styles.desc}>{item.Description}</Text>
      </View>

      <Text style={styles.sectionTitle}>Свободное время</Text>
      <FlatList
        data={slots}
        keyExtractor={(s) => String(s.ID)}
        contentContainerStyle={{ padding: 16, gap: 10 }}
        numColumns={2}
        columnWrapperStyle={{ gap: 10 }}
        renderItem={({ item: slot }) => (
          <TouchableOpacity
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
            <Text style={styles.slotText}>{formatSlot(slot.When)}</Text>
          </TouchableOpacity>
        )}
        ListEmptyComponent={<Text style={styles.meta}>Нет свободных слотов</Text>}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg },
  center: { flex: 1, alignItems: "center", justifyContent: "center", backgroundColor: colors.bg },
  info: { padding: 20, backgroundColor: colors.card, borderBottomWidth: 1, borderBottomColor: colors.border },
  emoji: { fontSize: 32 },
  name: { fontSize: 20, fontWeight: "800", color: colors.ink, marginTop: 6 },
  meta: { fontSize: 13, color: colors.muted, marginTop: 4 },
  price: { fontSize: 18, fontWeight: "800", color: colors.purpleDark, marginTop: 8 },
  desc: { fontSize: 13, color: colors.muted, marginTop: 8 },
  sectionTitle: { fontSize: 15, fontWeight: "700", color: colors.ink, paddingHorizontal: 16, paddingTop: 16 },
  slot: {
    flex: 1,
    backgroundColor: colors.card,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 12,
    padding: 12,
    alignItems: "center",
  },
  slotText: { fontSize: 13, fontWeight: "600", color: colors.ink },
});
