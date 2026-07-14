import { useEffect, useState } from "react";
import { View, Text, TouchableOpacity, StyleSheet, Alert } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { api } from "../api";
import { colors, radius, shadow } from "../theme";
import { useAuth } from "../AuthContext";
import Card from "../components/Card";
import GradientButton from "../components/GradientButton";

function formatSlot(iso) {
  const d = new Date(iso);
  return d.toLocaleString("ru-RU", {
    day: "2-digit",
    month: "long",
    hour: "2-digit",
    minute: "2-digit",
  });
}

const METHODS = [
  { id: "kaspi", label: "Kaspi Pay", icon: "wallet-outline", tone: "#ff4b53" },
  { id: "card", label: "Банковская карта", icon: "card-outline", tone: colors.blue },
];

export default function BookingScreen({ route, navigation }) {
  const { slotId, item, clinicName, slotWhen } = route.params;
  const { token } = useAuth();
  const [loading, setLoading] = useState(false);
  const [method, setMethod] = useState("kaspi");
  const [subscription, setSubscription] = useState(null);
  const [useSubscription, setUseSubscription] = useState(false);

  useEffect(() => {
    api.me(token).then((me) => {
      if (me.subscription && me.subscription.visitsLeft > 0) {
        setSubscription(me.subscription);
      }
    });
  }, [token]);

  async function confirmAndPay() {
    setLoading(true);
    try {
      const booking = await api.createBooking(slotId, useSubscription, token);
      if (!useSubscription) {
        await api.payBooking(booking.ID, token);
      }
      navigation.replace("Success", { item, clinicName, slotWhen, freeBySubscription: useSubscription });
    } catch (e) {
      Alert.alert("Не получилось", e.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <View style={styles.screen}>
      <Card style={styles.summary}>
        <Text style={styles.title}>Подтверждение записи</Text>
        <View style={styles.row}>
          <Ionicons name="person-outline" size={16} color={colors.muted} />
          <Text style={styles.rowText}>{item.Name}</Text>
        </View>
        <View style={styles.row}>
          <Ionicons name="business-outline" size={16} color={colors.muted} />
          <Text style={styles.rowText}>{clinicName}</Text>
        </View>
        <View style={styles.row}>
          <Ionicons name="calendar-outline" size={16} color={colors.muted} />
          <Text style={styles.rowText}>{formatSlot(slotWhen)}</Text>
        </View>
      </Card>

      {subscription && (
        <TouchableOpacity onPress={() => setUseSubscription((v) => !v)} activeOpacity={0.85}>
          <Card style={[styles.subCard, useSubscription && styles.subCardActive]}>
            <View style={{ flex: 1 }}>
              <Text style={styles.subTitle}>Списать визит по подписке «{subscription.planName}»</Text>
              <Text style={styles.subMeta}>Осталось визитов: {subscription.visitsLeft}</Text>
            </View>
            <Ionicons
              name={useSubscription ? "checkmark-circle" : "ellipse-outline"}
              size={24}
              color={useSubscription ? colors.teal : colors.faint}
            />
          </Card>
        </TouchableOpacity>
      )}

      {!useSubscription && (
        <>
          <Text style={styles.sectionLabel}>Способ оплаты</Text>
          <View style={{ gap: 10 }}>
            {METHODS.map((m) => (
              <TouchableOpacity key={m.id} activeOpacity={0.85} onPress={() => setMethod(m.id)}>
                <Card style={[styles.methodCard, method === m.id && styles.methodCardActive]}>
                  <View style={[styles.methodIcon, { backgroundColor: m.tone + "20" }]}>
                    <Ionicons name={m.icon} size={18} color={m.tone} />
                  </View>
                  <Text style={styles.methodLabel}>{m.label}</Text>
                  <Ionicons
                    name={method === m.id ? "radio-button-on" : "radio-button-off"}
                    size={20}
                    color={method === m.id ? colors.purple : colors.faint}
                  />
                </Card>
              </TouchableOpacity>
            ))}
          </View>
        </>
      )}

      <View style={styles.priceRow}>
        <Text style={styles.priceLabel}>К оплате</Text>
        <Text style={styles.price}>
          {useSubscription ? "0 ₸ (по подписке)" : `${item.Price.toLocaleString("ru-RU")} ₸`}
        </Text>
      </View>

      <Text style={styles.note}>
        Демо-режим: реальные банковские данные не запрашиваются и не списываются. Здесь будет
        подключён настоящий Kaspi Pay и другие банки.
      </Text>

      <GradientButton
        title={useSubscription ? "Подтвердить запись" : "Оплатить и подтвердить"}
        onPress={confirmAndPay}
        loading={loading}
        style={{ marginTop: 8 }}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bg, padding: 20, gap: 14 },
  summary: { gap: 10 },
  title: { fontSize: 17, fontWeight: "800", color: colors.ink, marginBottom: 2 },
  row: { flexDirection: "row", alignItems: "center", gap: 8 },
  rowText: { fontSize: 14, color: colors.ink },
  sectionLabel: { fontSize: 13, fontWeight: "700", color: colors.muted, marginTop: 4 },
  methodCard: { flexDirection: "row", alignItems: "center", gap: 12, padding: 14 },
  methodCardActive: { borderColor: colors.purple, borderWidth: 1.5 },
  methodIcon: { width: 36, height: 36, borderRadius: 10, alignItems: "center", justifyContent: "center" },
  methodLabel: { flex: 1, fontSize: 14, fontWeight: "600", color: colors.ink },
  subCard: { flexDirection: "row", alignItems: "center", gap: 12, borderColor: colors.teal, backgroundColor: "#effcfa" },
  subCardActive: { borderWidth: 1.5 },
  subTitle: { fontSize: 13.5, fontWeight: "700", color: colors.ink },
  subMeta: { fontSize: 12, color: colors.muted, marginTop: 3 },
  priceRow: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    backgroundColor: "#f1edff",
    borderRadius: radius.md,
    padding: 16,
    marginTop: 6,
  },
  priceLabel: { fontSize: 14, fontWeight: "600", color: colors.ink },
  price: { fontSize: 18, fontWeight: "800", color: colors.purpleDark },
  note: { fontSize: 11.5, color: colors.muted, textAlign: "center" },
});
