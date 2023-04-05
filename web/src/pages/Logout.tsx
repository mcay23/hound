function Logout() {
  localStorage.removeItem("isAuthenticated");
  localStorage.removeItem("username");
  window.location.reload();
  return <></>;
}

export default Logout;
