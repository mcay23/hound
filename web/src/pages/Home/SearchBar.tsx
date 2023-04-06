import { useState } from "react";
import { useNavigate } from "react-router-dom";
import "./SearchBar.css";

function SearchBar(props: any) {
  const navigate = useNavigate();
  const submitHandler = (event: any) => {
    event.preventDefault();
    if (searchQuery !== "") {
      navigate("/search?q=" + searchQuery);
      window.location.reload();
    }
  };
  const [searchQuery, setSearchQuery] = useState("");
  const onKeyChange = (event: any) => {
    setSearchQuery(event.target.value);
  };
  return (
    <form className="search-bar-container" onSubmit={submitHandler}>
      <input
        type="text"
        value={searchQuery}
        onChange={onKeyChange}
        placeholder="Search Anything..."
      />
      {props.type === "nav" ? "" : <button type="submit">GO</button>}
    </form>
  );
}

export default SearchBar;
