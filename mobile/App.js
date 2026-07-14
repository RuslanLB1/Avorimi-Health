import { StatusBar } from "expo-status-bar";
import { NavigationContainer } from "@react-navigation/native";
import { createNativeStackNavigator } from "@react-navigation/native-stack";
import { createBottomTabNavigator } from "@react-navigation/bottom-tabs";
import { Ionicons } from "@expo/vector-icons";
import { colors } from "./src/theme";
import { AuthProvider } from "./src/AuthContext";
import { FavoritesProvider } from "./src/FavoritesContext";

import ClinicsScreen from "./src/screens/ClinicsScreen";
import ClinicDetailScreen from "./src/screens/ClinicDetailScreen";
import CategoryScreen from "./src/screens/CategoryScreen";
import ItemScreen from "./src/screens/ItemScreen";
import BookingScreen from "./src/screens/BookingScreen";
import SuccessScreen from "./src/screens/SuccessScreen";
import LoginScreen from "./src/screens/LoginScreen";
import RegisterScreen from "./src/screens/RegisterScreen";
import AccountScreen from "./src/screens/AccountScreen";
import ResultsScreen from "./src/screens/ResultsScreen";
import SubscriptionsScreen from "./src/screens/SubscriptionsScreen";

const RootStack = createNativeStackNavigator();
const ClinicsStackNav = createNativeStackNavigator();
const AccountStackNav = createNativeStackNavigator();
const SubsStackNav = createNativeStackNavigator();
const Tab = createBottomTabNavigator();

const screenOptions = {
  headerStyle: { backgroundColor: colors.card },
  headerShadowVisible: false,
  headerTintColor: colors.ink,
  headerTitleStyle: { fontWeight: "700" },
};

function ClinicsStack() {
  return (
    <ClinicsStackNav.Navigator screenOptions={screenOptions}>
      <ClinicsStackNav.Screen name="Clinics" component={ClinicsScreen} options={{ headerShown: false }} />
      <ClinicsStackNav.Screen name="Clinic" component={ClinicDetailScreen} />
      <ClinicsStackNav.Screen name="Category" component={CategoryScreen} />
      <ClinicsStackNav.Screen name="Item" component={ItemScreen} options={{ title: "Специалист" }} />
      <ClinicsStackNav.Screen name="Booking" component={BookingScreen} options={{ title: "Запись" }} />
      <ClinicsStackNav.Screen name="Success" component={SuccessScreen} options={{ headerShown: false }} />
    </ClinicsStackNav.Navigator>
  );
}

function SubscriptionsStack() {
  return (
    <SubsStackNav.Navigator screenOptions={screenOptions}>
      <SubsStackNav.Screen name="SubscriptionsHome" component={SubscriptionsScreen} options={{ title: "Подписки" }} />
    </SubsStackNav.Navigator>
  );
}

function AccountStack() {
  return (
    <AccountStackNav.Navigator screenOptions={screenOptions}>
      <AccountStackNav.Screen name="AccountHome" component={AccountScreen} options={{ headerShown: false }} />
      <AccountStackNav.Screen name="Results" component={ResultsScreen} options={{ title: "Мои анализы" }} />
    </AccountStackNav.Navigator>
  );
}

function Tabs() {
  return (
    <Tab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: colors.purple,
        tabBarInactiveTintColor: colors.faint,
        tabBarStyle: { height: 62, paddingTop: 6, paddingBottom: 8 },
        tabBarLabelStyle: { fontSize: 11, fontWeight: "700" },
      }}
    >
      <Tab.Screen
        name="ClinicsTab"
        component={ClinicsStack}
        options={{ title: "Клиники", tabBarIcon: ({ color, size }) => <Ionicons name="location" size={size} color={color} /> }}
      />
      <Tab.Screen
        name="SubscriptionsTab"
        component={SubscriptionsStack}
        options={{ title: "Подписки", tabBarIcon: ({ color, size }) => <Ionicons name="ribbon" size={size} color={color} /> }}
      />
      <Tab.Screen
        name="AccountTab"
        component={AccountStack}
        options={{ title: "Профиль", tabBarIcon: ({ color, size }) => <Ionicons name="person" size={size} color={color} /> }}
      />
    </Tab.Navigator>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <FavoritesProvider>
        <NavigationContainer>
          <StatusBar style="dark" />
          <RootStack.Navigator screenOptions={screenOptions}>
            <RootStack.Screen name="Tabs" component={Tabs} options={{ headerShown: false }} />
            <RootStack.Screen name="Login" component={LoginScreen} options={{ title: "Вход", presentation: "modal" }} />
            <RootStack.Screen name="Register" component={RegisterScreen} options={{ title: "Регистрация", presentation: "modal" }} />
          </RootStack.Navigator>
        </NavigationContainer>
      </FavoritesProvider>
    </AuthProvider>
  );
}
