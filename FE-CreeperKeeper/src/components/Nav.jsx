import "../styles/nav.css";
import generateUUID from "../utils/uuid";

export default function Nav({ listItems }) {
  return (
    <nav className="main-nav">
      <ol>
        {listItems.map((i) => (
          <li key={generateUUID()}>
            <a onClick={i.callback !== null ? i.callback : null} href={i.link}>
              {i.item}
            </a>
          </li>
        ))}
      </ol>
    </nav>
  );
}
