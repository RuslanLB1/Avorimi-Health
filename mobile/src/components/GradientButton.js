import { TouchableOpacity, Text, StyleSheet, ActivityIndicator } from "react-native";
import { LinearGradient } from "expo-linear-gradient";
import { gradients, radius } from "../theme";

export default function GradientButton({ title, onPress, loading, disabled, style, colors, textColor }) {
  const color = textColor || "#fff";
  return (
    <TouchableOpacity
      activeOpacity={0.85}
      onPress={onPress}
      disabled={disabled || loading}
      style={[styles.wrap, style, (disabled || loading) && styles.disabled]}
    >
      <LinearGradient
        colors={colors || gradients.brand}
        start={{ x: 0, y: 0 }}
        end={{ x: 1, y: 1 }}
        style={styles.gradient}
      >
        {loading ? <ActivityIndicator color={color} /> : <Text style={[styles.text, { color }]}>{title}</Text>}
      </LinearGradient>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  wrap: { borderRadius: radius.md, overflow: "hidden" },
  gradient: { paddingVertical: 16, alignItems: "center", justifyContent: "center" },
  text: { fontWeight: "700", fontSize: 15 },
  disabled: { opacity: 0.6 },
});
