import { ref } from "vue";

const THEME_KEY = "theme";
const isDark = ref(false);

/**
 * 在应用 mount 之前调用，避免首帧闪烁（FOUC）
 * 同步读取 localStorage，无值时检测系统偏好
 */
export function initTheme() {
  const saved = localStorage.getItem(THEME_KEY);
  if (saved) {
    isDark.value = saved === "dark";
  } else {
    isDark.value = window.matchMedia("(prefers-color-scheme: dark)").matches;
  }
  applyTheme();
}

function applyTheme() {
  const html = document.documentElement;
  if (isDark.value) {
    html.classList.add("dark");
  } else {
    html.classList.remove("dark");
  }
}

/**
 * 主题切换 composable，全局共享同一份 isDark 状态
 */
export function useTheme() {
  const toggle = () => {
    isDark.value = !isDark.value;
    applyTheme();
    localStorage.setItem(THEME_KEY, isDark.value ? "dark" : "light");
  };

  return { isDark, toggle };
}
