import { useEffect, useState } from "react";
import { View, Text, ScrollView, TouchableOpacity, StyleSheet, Alert } from "react-native";
import { LinearGradient } from "expo-linear-gradient";
import { Ionicons } from "@expo/vector-icons";
import { api } from "../api";
import { colors, gradients, radius, shadow } from "../theme";
import { useAuth } from "../AuthContext";
import GradientButton from "../components/GradientButton";
import Card from "../components/Card";

export default function SubscriptionsScreen({ navigation }) {
  const { token, user } = useAuth();
  const [plans, setPlans] = useState(null);
  const [activePlanId, setActivePlanId] = useState(null);
  const [subscribing, setSubscribing] = useState(null);

  useEffect(() => {
    api.plans().then(setPlans);
    if (token) {
      api.me(token).then((me) => {
        if (me.subscription) setActivePlanId(me.subscription.planName);
      });
    }
  }, [token]);

  async function subscribe(plan) {
    if (!user) {
      navigation.navigate("Login");
      return;
    }
    setSubscribing(plan.ID);
    try {
      await api.subscribe(plan.ID, token);
      Alert.alert("Готово", `Подписка «${plan.Name}» активирована`);
      const me = await api.me(token);
      if (me.subscription) setActivePlanId(me.subscription.planName);
    } catch (e) {
      Alert.alert("Не получилось", e.message);
    } finally {
      setSubscribing(null);
    }
  }

  return (
    <ScrollView style={styles.screen} contentContainerStyle={{ padding: 20, gap: 16 }}>
      <Text style={styles.intro}>
        Безлимитные визиты по фиксированной цене в месяц — выгоднее разовой оплаты каждого приёма.
      </Text>
      {(plans || []).map((plan) => (
        <View key={plan.ID}>
          {plan.Highlight ? (
            <LinearGradient colors={gradients.brand} style={[styles.card, styles.cardHighlight]}>
              <PlanContent plan={plan} highlighted onSubscribe={() => subscribe(plan)} loading={subscribing === plan.ID} active={activePlanId === plan.Name} />
            </LinearGradient>
          ) : (
            <Card style={styles.card}>
              <PlanContent plan={plan} onSubscribe={() => subscribe(plan)} loading={subscribing === plan.ID} active={activePlanId === plan.Name} />
            </Card>
          )}
        </View>
      ))}
    </ScrollView>
  );
}

function PlanContent({ plan, highlighted, onSubscribe, loading, active }) {
  const textColor = highlighted ? "#fff" : colors.ink;
  const mutedColor = highlighted ? "rgba(255,255,255,0.85)" : colors.muted;
  return (
    <>
      {plan.Highlight && (
        <View style={styles.badge}>
          <Ionicons name="star" size={11} color="#fff" />
          <Text style={styles.badgeText}>Популярный</Text>
        </View>
      )}
      <Text style={[styles.planName, { color: textColor }]}>{plan.Name}</Text>
      <Text style={[styles.planPrice, { color: textColor }]}>{plan.Price}</Text>
      <Text style={[styles.planDesc, { color: mutedColor }]}>{plan.Description}</Text>
      {active ? (
        <View style={styles.activePill}>
          <Ionicons name="checkmark-circle" size={16} color={highlighted ? "#fff" : colors.teal} />
          <Text style={{ color: highlighted ? "#fff" : colors.teal, fontWeight: "700", fontSize: 13 }}>Активна</Text>
        </View>
      ) : highlighted ? (
        <TouchableOpacity activeOpacity={0.85} onPress={onSubscribe} style={styles.whiteButton} disabled={loading}>
          <Text style={styles.whiteButtonText}>{loading ? "..." : "Оформить"}</Text>
        </TouchableOpacity>
      ) : (
        <GradientButton title="Оформить" onPress={onSubscribe} loading={loading} style={{ marginTop: 14 }} />
      )}
    </>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg },
  intro: { fontSize: 13.5, color: colors.muted, textAlign: "center", lineHeight: 20 },
  card: { padding: 20, borderRadius: radius.lg, ...shadow.card },
  cardHighlight: { borderWidth: 0 },
  badge: {
    flexDirection: "row",
    alignItems: "center",
    gap: 4,
    backgroundColor: "rgba(255,255,255,0.25)",
    alignSelf: "flex-start",
    paddingVertical: 4,
    paddingHorizontal: 10,
    borderRadius: radius.pill,
    marginBottom: 10,
  },
  badgeText: { color: "#fff", fontSize: 11, fontWeight: "700" },
  planName: { fontSize: 18, fontWeight: "800" },
  planPrice: { fontSize: 24, fontWeight: "800", marginTop: 6 },
  planDesc: { fontSize: 13, marginTop: 10, lineHeight: 19 },
  activePill: { flexDirection: "row", alignItems: "center", gap: 6, marginTop: 14 },
  whiteButton: {
    backgroundColor: "#fff",
    borderRadius: radius.md,
    paddingVertical: 14,
    alignItems: "center",
    marginTop: 14,
  },
  whiteButtonText: { color: colors.purpleDark, fontWeight: "800", fontSize: 15 },
});
