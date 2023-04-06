import "./Footer.css";

function Footer(props: any) {
  return (
    <div className="footer-main-section">
      <div className="footer-logo-header">Powered By</div>
      <div className="footer-logo-container">
        <img
          src="https://www.themoviedb.org/assets/2/v4/logos/v2/blue_square_1-5bdc75aaebeb75dc7ae79426ddd9be3b2be1e342510f8202baf6bffa71d7f5c4.svg"
          alt="tmdb-logo"
          id="tmdb-logo"
        />
        <img
          src="https://upload.wikimedia.org/wikipedia/commons/1/19/IGDB_logo.svg"
          alt="igdb-logo"
          id="igdb-logo"
        />
      </div>
    </div>
  );
}

export default Footer;
