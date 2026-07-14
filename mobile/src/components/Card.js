import { View, StyleSheet } from "react-native";
import { colors, radius, shadow } from "../theme";

export default function Card({ children, style, noPad }) {
  return <View style={[styles.card, !noPad && styles.pad, style]}>{children}</View>;
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: colors.card,
    borderRadius: radius.lg,
    borderWidth: 1,
    borderColor: colors.border,
    ...shadow.soft,
  },
  pad: { padding: 16 },
});
