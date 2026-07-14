import { View, Text, Image, TouchableOpacity, StyleSheet } from "react-native";
import { LinearGradient } from "expo-linear-gradient";
import { Ionicons } from "@expo/vector-icons";
import { gradients, radius } from "../theme";
import GradientButton from "../components/GradientButton";

export default function HomeScreen({ navigation }) {
  return (
    <LinearGradient colors={gradients.brand} style={styles.screen}>
      <View style={styles.top}>
        <View style={styles.logoWrap}>
          <Image source={require("../../assets/logo.png")} style={styles.logo} />
        </View>
        <Text style={styles.title}>Avorimi Health</Text>
        <Text style={styles.tagline}>Клиники, врачи и запись на приём — в одном приложении</Text>
      </View>

      <View style={styles.features}>
        <View style={styles.featureRow}>
          <Ionicons name="location" size={16} color="#fff" />
          <Text style={styles.featureText}>Клиники рядом с вами</Text>
        </View>
        <View style={styles.featureRow}>
          <Ionicons name="calendar" size={16} color="#fff" />
          <Text style={styles.featureText}>Свободное время онлайн</Text>
        </View>
        <View style={styles.featureRow}>
          <Ionicons name="shield-checkmark" size={16} color="#fff" />
          <Text style={styles.featureText}>Без регистрации — до момента записи</Text>
        </View>
      </View>

      <View style={styles.actions}>
        <GradientButton
          title="📍 Найти клинику рядом"
          colors={["#ffffff", "#ffffff"]}
          textColor="#6238e0"
          onPress={() => navigation.replace("Tabs", { screen: "ClinicsTab" })}
        />
        <View style={{ height: 12 }} />
        <TouchableOpacity
          onPress={() => navigation.replace("Tabs", { screen: "SubscriptionsTab" })}
          style={styles.secondaryBtn}
          activeOpacity={0.8}
        >
          <Text style={styles.secondaryText}>Посмотреть подписки</Text>
        </TouchableOpacity>
      </View>
    </LinearGradient>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, justifyContent: "space-between", padding: 28, paddingTop: 80, paddingBottom: 48 },
  top: { alignItems: "center" },
  logoWrap: {
    width: 96,
    height: 96,
    borderRadius: radius.xl,
    backgroundColor: "rgba(255,255,255,0.15)",
    alignItems: "center",
    justifyContent: "center",
    marginBottom: 20,
    overflow: "hidden",
  },
  logo: { width: 76, height: 76, borderRadius: radius.lg },
  title: { color: "#fff", fontSize: 30, fontWeight: "800", letterSpacing: 0.3 },
  tagline: { color: "rgba(255,255,255,0.88)", fontSize: 14.5, textAlign: "center", marginTop: 10, lineHeight: 21, maxWidth: 280 },
  features: { gap: 14 },
  featureRow: { flexDirection: "row", alignItems: "center", gap: 10, alignSelf: "center" },
  featureText: { color: "rgba(255,255,255,0.92)", fontSize: 13.5, fontWeight: "600" },
  actions: {},
  secondaryBtn: { alignItems: "center", paddingVertical: 12 },
  secondaryText: { color: "rgba(255,255,255,0.85)", fontWeight: "700", fontSize: 14 },
});
