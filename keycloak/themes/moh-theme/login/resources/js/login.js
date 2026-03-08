document.addEventListener("DOMContentLoaded", function () {
  const toggle = document.querySelector("[data-password-toggle]");
  const password = document.getElementById("password");

  if (toggle && password) {
    toggle.addEventListener("click", function () {
      const isPassword = password.getAttribute("type") === "password";
      password.setAttribute("type", isPassword ? "text" : "password");
      toggle.textContent = isPassword ? "Hide" : "Show";
    });
  }
});