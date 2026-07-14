import { View, Text, StyleSheet } from "react-native";
import { colors } from "../theme";

export default function EmptyState({ icon = "🔍", title, subtitle }) {
  return (
    <View style={styles.wrap}>
      <Text style={styles.icon}>{icon}</Text>
      {title ? <Text style={styles.title}>{title}</Text> : null}
      {subtitle ? <Text style={styles.subtitle}>{subtitle}</Text> : null}
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { alignItems: "center", justifyContent: "center", padding: 40, gap: 8 },
  icon: { fontSize: 40 },
  title: { fontSize: 15, fontWeight: "700", color: colors.ink, textAlign: "center" },
  subtitle: { fontSize: 13, color: colors.muted, textAlign: "center" },
});
