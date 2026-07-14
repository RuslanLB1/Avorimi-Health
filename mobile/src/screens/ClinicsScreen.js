import { useEffect, useState, useCallback, useMemo } from "react";
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  RefreshControl,
  TextInput,
} from "react-native";
import { LinearGradient } from "expo-linear-gradient";
import { Ionicons } from "@expo/vector-icons";
import * as Location from "expo-location";
import { api } from "../api";
import { colors, gradients, radius, shadow } from "../theme";
import { useAuth } from "../AuthContext";
import { useFavorites } from "../FavoritesContext";
import { SkeletonCard } from "../components/Skeleton";
import EmptyState from "../components/EmptyState";

export default function ClinicsScreen({ navigation }) {
  const { user } = useAuth();
  const { isFavorite, toggle } = useFavorites();
  const [clinics, setClinics] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [locationNote, setLocationNote] = useState("");
  const [query, setQuery] = useState("");
  const [onlyFavorites, setOnlyFavorites] = useState(false);

  const load = useCallback(async () => {
    let lat, lng;
    try {
      const { status } = await Location.requestForegroundPermissionsAsync();
      if (status === "granted") {
        const pos = await Location.getCurrentPositionAsync({});
        lat = pos.coords.latitude;
        lng = pos.coords.longitude;
        setLocationNote("По расстоянию от вас");
      } else {
        setLocationNote("Включите геолокацию — покажем ближайшие");
      }
    } catch {
      setLocationNote("Не удалось определить геолокацию");
    }
    const data = await api.clinics(lat, lng);
    setClinics(data);
    setLoading(false);
    setRefreshing(false);
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const filtered = useMemo(() => {
    let list = clinics;
    if (onlyFavorites) list = list.filter((c) => isFavorite(c.ID));
    if (query.trim()) {
      const q = query.trim().toLowerCase();
      list = list.filter(
        (c) => c.Name.toLowerCase().includes(q) || c.Address.toLowerCase().includes(q)
      );
    }
    return list;
  }, [clinics, query, onlyFavorites, isFavorite]);

  return (
    <View style={styles.screen}>
      <LinearGradient colors={gradients.brand} style={styles.header}>
        <View style={styles.headerTop}>
          <View>
            <Text style={styles.eyebrow}>AVORIMI HEALTH</Text>
            <Text style={styles.title}>Клиники рядом</Text>
          </View>
          <TouchableOpacity
            style={styles.profileDot}
            onPress={() => navigation.navigate(user ? "AccountTab" : "Login")}
          >
            <Ionicons name={user ? "person" : "log-in-outline"} size={20} color="#fff" />
          </TouchableOpacity>
        </View>
        <Text style={styles.subtitle}>{locationNote}</Text>

        <View style={styles.searchBox}>
          <Ionicons name="search" size={18} color={colors.muted} />
          <TextInput
            style={styles.searchInput}
            placeholder="Название или адрес клиники"
            placeholderTextColor={colors.faint}
            value={query}
            onChangeText={setQuery}
          />
        </View>
      </LinearGradient>

      <View style={styles.filterRow}>
        <TouchableOpacity
          style={[styles.chip, !onlyFavorites && styles.chipActive]}
          onPress={() => setOnlyFavorites(false)}
        >
          <Text style={[styles.chipText, !onlyFavorites && styles.chipTextActive]}>Все</Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.chip, onlyFavorites && styles.chipActive]}
          onPress={() => setOnlyFavorites(true)}
        >
          <Ionicons name="heart" size={13} color={onlyFavorites ? "#fff" : colors.purple} />
          <Text style={[styles.chipText, onlyFavorites && styles.chipTextActive]}>Избранное</Text>
        </TouchableOpacity>
      </View>

      <FlatList
        data={loading ? Array.from({ length: 5 }) : filtered}
        keyExtractor={(c, i) => (c ? String(c.ID) : String(i))}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            tintColor={colors.purple}
            onRefresh={() => {
              setRefreshing(true);
              load();
            }}
          />
        }
        contentContainerStyle={{ padding: 16, gap: 12, paddingBottom: 32 }}
        ListEmptyComponent={
          !loading && (
            <EmptyState
              icon={onlyFavorites ? "💜" : "🏥"}
              title={onlyFavorites ? "Пока нет избранных клиник" : "Ничего не найдено"}
              subtitle={onlyFavorites ? "Нажми на сердечко у клиники, чтобы сохранить" : "Попробуй изменить запрос"}
            />
          )
        }
        renderItem={({ item }) =>
          loading ? (
            <SkeletonCard />
          ) : (
            <TouchableOpacity
              activeOpacity={0.8}
              style={styles.card}
              onPress={() => navigation.navigate("Clinic", { clinicId: item.ID, name: item.Name })}
            >
              <View style={styles.emojiWrap}>
                <Text style={styles.emoji}>{item.Emoji}</Text>
              </View>
              <View style={{ flex: 1 }}>
                <Text style={styles.cardTitle} numberOfLines={1}>{item.Name}</Text>
                <Text style={styles.cardMeta} numberOfLines={1}>{item.Address}</Text>
                <View style={styles.metaRow}>
                  <Ionicons name="star" size={12} color={colors.gold} />
                  <Text style={styles.metaText}>{item.Rating}</Text>
                  <Text style={styles.dot}>·</Text>
                  <Text style={styles.metaText}>{item.itemCount} специалистов</Text>
                  {item.distanceKm ? (
                    <>
                      <Text style={styles.dot}>·</Text>
                      <Text style={styles.metaText}>{item.distanceKm.toFixed(1)} км</Text>
                    </>
                  ) : null}
                </View>
              </View>
              <TouchableOpacity hitSlop={10} onPress={() => toggle(item.ID)}>
                <Ionicons
                  name={isFavorite(item.ID) ? "heart" : "heart-outline"}
                  size={22}
                  color={isFavorite(item.ID) ? colors.purple : colors.faint}
                />
              </TouchableOpacity>
            </TouchableOpacity>
          )
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg },
  header: {
    paddingTop: 58,
    paddingHorizontal: 20,
    paddingBottom: 22,
    borderBottomLeftRadius: radius.xl,
    borderBottomRightRadius: radius.xl,
  },
  headerTop: { flexDirection: "row", justifyContent: "space-between", alignItems: "flex-start" },
  eyebrow: { color: "rgba(255,255,255,0.75)", fontSize: 11, fontWeight: "700", letterSpacing: 1 },
  title: { color: "#fff", fontSize: 24, fontWeight: "800", marginTop: 2 },
  subtitle: { color: "rgba(255,255,255,0.85)", fontSize: 12.5, marginTop: 8 },
  profileDot: {
    width: 38,
    height: 38,
    borderRadius: 19,
    backgroundColor: "rgba(255,255,255,0.2)",
    alignItems: "center",
    justifyContent: "center",
  },
  searchBox: {
    flexDirection: "row",
    alignItems: "center",
    gap: 8,
    backgroundColor: "#fff",
    borderRadius: radius.md,
    paddingHorizontal: 14,
    paddingVertical: 12,
    marginTop: 16,
    ...shadow.card,
  },
  searchInput: { flex: 1, fontSize: 14, color: colors.ink },
  filterRow: { flexDirection: "row", gap: 8, paddingHorizontal: 16, paddingTop: 14 },
  chip: {
    flexDirection: "row",
    alignItems: "center",
    gap: 5,
    paddingVertical: 7,
    paddingHorizontal: 14,
    borderRadius: radius.pill,
    backgroundColor: colors.card,
    borderWidth: 1,
    borderColor: colors.border,
  },
  chipActive: { backgroundColor: colors.purple, borderColor: colors.purple },
  chipText: { fontSize: 13, fontWeight: "600", color: colors.ink },
  chipTextActive: { color: "#fff" },
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
    width: 48,
    height: 48,
    borderRadius: radius.md,
    backgroundColor: "#f1edff",
    alignItems: "center",
    justifyContent: "center",
  },
  emoji: { fontSize: 24 },
  cardTitle: { fontSize: 15, fontWeight: "700", color: colors.ink },
  cardMeta: { fontSize: 12, color: colors.muted, marginTop: 2 },
  metaRow: { flexDirection: "row", alignItems: "center", gap: 4, marginTop: 6 },
  metaText: { fontSize: 12, color: colors.muted, fontWeight: "600" },
  dot: { color: colors.faint },
});
