import { useCallback, useState } from "react";
import { View, Text, FlatList, StyleSheet } from "react-native";
import { useFocusEffect } from "@react-navigation/native";
import { Ionicons } from "@expo/vector-icons";
import { api } from "../api";
import { colors, radius, shadow } from "../theme";
import { useAuth } from "../AuthContext";
import Card from "../components/Card";
import Badge from "../components/Badge";
import EmptyState from "../components/EmptyState";
import { SkeletonCard } from "../components/Skeleton";

export default function ResultsScreen() {
  const { token } = useAuth();
  const [results, setResults] = useState(null);

  useFocusEffect(
    useCallback(() => {
      if (token) api.results(token).then(setResults);
    }, [token])
  );

  return (
    <FlatList
      style={styles.screen}
      data={results || Array.from({ length: 3 })}
      keyExtractor={(r, i) => (r ? String(r.booking.ID) : String(i))}
      contentContainerStyle={{ padding: 16, gap: 12 }}
      ListEmptyComponent={
        results && (
          <EmptyState icon="🧪" title="Пока нет анализов" subtitle="Записи на анализы и диагностику появятся здесь" />
        )
      }
      renderItem={({ item }) =>
        !results ? (
          <SkeletonCard />
        ) : (
          <Card style={styles.card}>
            <View style={styles.iconWrap}>
              <Ionicons name="flask-outline" size={20} color={colors.purple} />
            </View>
            <View style={{ flex: 1 }}>
              <Text style={styles.title}>{item.item?.Category}</Text>
              <Text style={styles.meta}>
                {new Date(item.slot?.When).toLocaleDateString("ru-RU", { day: "2-digit", month: "long" })}
              </Text>
            </View>
            <Badge
              label={item.ready ? "Готово" : "Ожидается"}
              tone={item.ready ? "teal" : "gold"}
            />
          </Card>
        )
      }
    />
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg },
  card: { flexDirection: "row", alignItems: "center", gap: 12 },
  iconWrap: {
    width: 40,
    height: 40,
    borderRadius: radius.md,
    backgroundColor: "#f1edff",
    alignItems: "center",
    justifyContent: "center",
  },
  title: { fontSize: 14.5, fontWeight: "700", color: colors.ink },
  meta: { fontSize: 12, color: colors.muted, marginTop: 3 },
});
