import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { colors } from "../theme";

export default function SuccessScreen({ route, navigation }) {
  const { item, clinicName } = route.params;

  return (
    <View style={styles.screen}>
      <Text style={styles.icon}>✅</Text>
      <Text style={styles.title}>Вы записаны!</Text>
      <Text style={styles.meta}>{item.Name} · {clinicName}</Text>
      <TouchableOpacity
        style={styles.button}
        onPress={() => navigation.popToTop()}
      >
        <Text style={styles.buttonText}>На главную</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg, alignItems: "center", justifyContent: "center", padding: 20 },
  icon: { fontSize: 56, marginBottom: 12 },
  title: { fontSize: 22, fontWeight: "800", color: colors.ink },
  meta: { fontSize: 14, color: colors.muted, marginTop: 8, textAlign: "center" },
  button: { backgroundColor: colors.purple, borderRadius: 14, padding: 16, paddingHorizontal: 28, marginTop: 24 },
  buttonText: { color: "#fff", fontWeight: "700", fontSize: 15 },
});
