import { useEffect, useState, useCallback } from "react";
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  RefreshControl,
} from "react-native";
import * as Location from "expo-location";
import { api } from "../api";
import { colors } from "../theme";
import { useAuth } from "../AuthContext";

export default function ClinicsScreen({ navigation }) {
  const { user, signOut } = useAuth();
  const [clinics, setClinics] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [locationNote, setLocationNote] = useState("");

  const load = useCallback(async () => {
    let lat, lng;
    try {
      const { status } = await Location.requestForegroundPermissionsAsync();
      if (status === "granted") {
        const pos = await Location.getCurrentPositionAsync({});
        lat = pos.coords.latitude;
        lng = pos.coords.longitude;
        setLocationNote("Отсортировано по расстоянию от вас");
      } else {
        setLocationNote("Включите геолокацию, чтобы видеть клиники рядом");
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

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={colors.purple} size="large" />
      </View>
    );
  }

  return (
    <View style={styles.screen}>
      <View style={styles.header}>
        <View>
          <Text style={styles.title}>Клиники рядом</Text>
          <Text style={styles.subtitle}>{locationNote}</Text>
        </View>
        <TouchableOpacity
          onPress={() => (user ? signOut() : navigation.navigate("Login"))}
        >
          <Text style={styles.link}>{user ? user.fullName.split(" ")[0] : "Войти"}</Text>
        </TouchableOpacity>
      </View>

      <FlatList
        data={clinics}
        keyExtractor={(c) => String(c.ID)}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={() => {
              setRefreshing(true);
              load();
            }}
          />
        }
        contentContainerStyle={{ padding: 16, gap: 12 }}
        renderItem={({ item }) => (
          <TouchableOpacity
            style={styles.card}
            onPress={() => navigation.navigate("Clinic", { clinicId: item.ID, name: item.Name })}
          >
            <Text style={styles.emoji}>{item.Emoji}</Text>
            <View style={{ flex: 1 }}>
              <Text style={styles.cardTitle}>{item.Name}</Text>
              <Text style={styles.cardMeta}>{item.Address}</Text>
              <Text style={styles.cardMeta}>
                ⭐ {item.Rating} · {item.itemCount} специалистов
                {item.distanceKm ? ` · ${item.distanceKm.toFixed(1)} км` : ""}
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
  header: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    padding: 16,
    paddingTop: 56,
    backgroundColor: colors.card,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  title: { fontSize: 20, fontWeight: "800", color: colors.ink },
  subtitle: { fontSize: 12, color: colors.muted, marginTop: 2 },
  link: { color: colors.purple, fontWeight: "700" },
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
  emoji: { fontSize: 28 },
  cardTitle: { fontSize: 16, fontWeight: "700", color: colors.ink },
  cardMeta: { fontSize: 12.5, color: colors.muted, marginTop: 2 },
});
