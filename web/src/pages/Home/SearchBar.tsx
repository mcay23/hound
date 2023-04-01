import { useState } from "react";
import { useNavigate } from "react-router-dom";
import "./SearchBar.css";

function SearchBar() {
  const navigate = useNavigate();
  const submitHandler = (event: any) => {
    event.preventDefault();
    console.log(searchQuery, "query");
    if (searchQuery !== "") {
      navigate("/search?q=" + searchQuery);
      window.location.reload();
    }
  };
  const [searchQuery, setSearchQuery] = useState("");
  const onKeyChange = (event: any) => {
    console.log(event.target.value);
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
      <button type="submit">
        {/* <SearchIcon fontSize="inherit" /> */}
        GO
      </button>
    </form>
  );
}

export default SearchBar;
