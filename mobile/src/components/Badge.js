import { View, Text, StyleSheet } from "react-native";
import { colors, radius } from "../theme";

const tones = {
  purple: { bg: "#efeaff", fg: colors.purpleDark },
  teal: { bg: "#e2f9f6", fg: "#0e8f81" },
  gold: { bg: "#fff3de", fg: "#a8690a" },
  muted: { bg: "#f0f0f5", fg: colors.muted },
};

export default function Badge({ label, tone = "purple", icon }) {
  const t = tones[tone] || tones.purple;
  return (
    <View style={[styles.badge, { backgroundColor: t.bg }]}>
      {icon}
      <Text style={[styles.text, { color: t.fg }]}>{label}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  badge: {
    flexDirection: "row",
    alignItems: "center",
    gap: 4,
    paddingVertical: 5,
    paddingHorizontal: 10,
    borderRadius: radius.pill,
    alignSelf: "flex-start",
  },
  text: { fontSize: 12, fontWeight: "700" },
});
