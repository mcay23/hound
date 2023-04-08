function Logout() {
  localStorage.removeItem("isAuthenticated");
  localStorage.removeItem("username");
  localStorage.removeItem("token");
  window.location.reload();
  return <></>;
}

export default Logout;
