function Logout() {
  localStorage.removeItem("isAuthenticated");
  window.location.reload();
  return <></>;
}

export default Logout;
