import { Text } from "react-native";
import { StatusBar } from "expo-status-bar";
import { NavigationContainer } from "@react-navigation/native";
import { createNativeStackNavigator } from "@react-navigation/native-stack";
import { createBottomTabNavigator } from "@react-navigation/bottom-tabs";
import { colors } from "./src/theme";
import { AuthProvider } from "./src/AuthContext";

import ClinicsScreen from "./src/screens/ClinicsScreen";
import ClinicDetailScreen from "./src/screens/ClinicDetailScreen";
import CategoryScreen from "./src/screens/CategoryScreen";
import ItemScreen from "./src/screens/ItemScreen";
import BookingScreen from "./src/screens/BookingScreen";
import SuccessScreen from "./src/screens/SuccessScreen";
import LoginScreen from "./src/screens/LoginScreen";
import RegisterScreen from "./src/screens/RegisterScreen";
import AccountScreen from "./src/screens/AccountScreen";

const RootStack = createNativeStackNavigator();
const ClinicsStackNav = createNativeStackNavigator();
const AccountStackNav = createNativeStackNavigator();
const Tab = createBottomTabNavigator();

const screenOptions = {
  headerStyle: { backgroundColor: colors.card },
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

function AccountStack() {
  return (
    <AccountStackNav.Navigator screenOptions={screenOptions}>
      <AccountStackNav.Screen name="Account" component={AccountScreen} options={{ title: "Мои записи" }} />
    </AccountStackNav.Navigator>
  );
}

function Tabs() {
  return (
    <Tab.Navigator screenOptions={{ headerShown: false, tabBarActiveTintColor: colors.purple }}>
      <Tab.Screen
        name="ClinicsTab"
        component={ClinicsStack}
        options={{ title: "Клиники", tabBarIcon: ({ size }) => <Text style={{ fontSize: size }}>📍</Text> }}
      />
      <Tab.Screen
        name="AccountTab"
        component={AccountStack}
        options={{ title: "Профиль", tabBarIcon: ({ size }) => <Text style={{ fontSize: size }}>👤</Text> }}
      />
    </Tab.Navigator>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <NavigationContainer>
        <StatusBar style="dark" />
        <RootStack.Navigator screenOptions={screenOptions}>
          <RootStack.Screen name="Tabs" component={Tabs} options={{ headerShown: false }} />
          <RootStack.Screen
            name="Login"
            component={LoginScreen}
            options={{ title: "Вход", presentation: "modal" }}
          />
          <RootStack.Screen
            name="Register"
            component={RegisterScreen}
            options={{ title: "Регистрация", presentation: "modal" }}
          />
        </RootStack.Navigator>
      </NavigationContainer>
    </AuthProvider>
  );
}
