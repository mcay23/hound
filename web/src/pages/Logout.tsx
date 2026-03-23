function Logout() {
  localStorage.removeItem("isAuthenticated");
  localStorage.removeItem("username");
  localStorage.removeItem("role");
  window.location.reload();
  return <></>;
}

export default Logout;
