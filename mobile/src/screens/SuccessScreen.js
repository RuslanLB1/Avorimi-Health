import { View, Text, StyleSheet } from "react-native";
import { LinearGradient } from "expo-linear-gradient";
import { Ionicons } from "@expo/vector-icons";
import { colors, gradients } from "../theme";
import GradientButton from "../components/GradientButton";

export default function SuccessScreen({ route, navigation }) {
  const { item, clinicName, freeBySubscription } = route.params;

  return (
    <View style={styles.screen}>
      <LinearGradient colors={gradients.teal} style={styles.icon}>
        <Ionicons name="checkmark" size={40} color="#fff" />
      </LinearGradient>
      <Text style={styles.title}>Вы записаны!</Text>
      <Text style={styles.meta}>{item.Name} · {clinicName}</Text>
      {freeBySubscription ? <Text style={styles.subNote}>Списано по подписке</Text> : null}
      <GradientButton
        title="На главную"
        onPress={() => navigation.popToTop()}
        style={{ marginTop: 28, width: 220 }}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg, alignItems: "center", justifyContent: "center", padding: 20 },
  icon: { width: 84, height: 84, borderRadius: 42, alignItems: "center", justifyContent: "center", marginBottom: 16 },
  title: { fontSize: 22, fontWeight: "800", color: colors.ink },
  meta: { fontSize: 14, color: colors.muted, marginTop: 8, textAlign: "center" },
  subNote: { fontSize: 12.5, color: colors.teal, fontWeight: "700", marginTop: 6 },
});
