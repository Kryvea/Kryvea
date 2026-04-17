type KryveaShadow = "admin" | "user" | "password_expired" | false;

export function getKryveaShadow(): KryveaShadow {
  const cookies = document.cookie.split("; ");
  const kryveaShadowCookie = cookies.find(cookie => cookie.startsWith("kryvea_shadow="));
  return kryveaShadowCookie ? (kryveaShadowCookie.split("=")[1] as KryveaShadow) : false;
}
