import { useEffect, useRef } from "react";
import { Animated, View, StyleSheet } from "react-native";
import { radius } from "../theme";

export function SkeletonBlock({ width = "100%", height = 16, style, round = radius.sm }) {
  const opacity = useRef(new Animated.Value(0.5)).current;

  useEffect(() => {
    const loop = Animated.loop(
      Animated.sequence([
        Animated.timing(opacity, { toValue: 1, duration: 650, useNativeDriver: true }),
        Animated.timing(opacity, { toValue: 0.5, duration: 650, useNativeDriver: true }),
      ])
    );
    loop.start();
    return () => loop.stop();
  }, [opacity]);

  return (
    <Animated.View
      style={[
        { width, height, borderRadius: round, backgroundColor: "#e5e1f7", opacity },
        style,
      ]}
    />
  );
}

export function SkeletonCard() {
  return (
    <View style={styles.card}>
      <SkeletonBlock width={44} height={44} round={22} />
      <View style={{ flex: 1, gap: 8 }}>
        <SkeletonBlock width="70%" height={14} />
        <SkeletonBlock width="45%" height={11} />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    flexDirection: "row",
    gap: 14,
    alignItems: "center",
    backgroundColor: "#fff",
    borderRadius: radius.lg,
    padding: 16,
    borderWidth: 1,
    borderColor: "#e7e4f5",
  },
});
