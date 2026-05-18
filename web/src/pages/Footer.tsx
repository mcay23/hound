import { GitHub } from "@mui/icons-material";
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
        <a
          href="https://github.com/Hound-Media-Server/hound"
          target="_blank"
          rel="noopener noreferrer"
        >
          <GitHub sx={{ color: "#FFFFFF", fontSize: "100px" }} />
        </a>
        <a
          href="https://reddit.com/r/HoundMediaServer"
          target="_blank"
          rel="noopener noreferrer"
        >
          <img
            src="https://upload.wikimedia.org/wikipedia/en/b/bd/Reddit_Logo_Icon.svg"
            alt="reddit-logo"
            id="reddit-logo"
          />
        </a>
      </div>
    </div>
  );
}

export default Footer;
