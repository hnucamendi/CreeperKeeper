import React from "react";
import "../styles/footer.css";

export default function Footer() {
  return (
    <footer className="main-footer">
      <div className="footer-section">
        <ul className="footer-links">
          <li>
            <a href="/about">About Us</a>
          </li>
          <li>
            <a href="/contact">Contact</a>
          </li>
          <li>
            <a href="/privacy">Privacy Policy</a>
          </li>
        </ul>
      </div>
      <div className="footer-section">
        <ul className="footer-socials">
          <li>
            <a
              href="https://facebook.com"
              target="_blank"
              rel="noopener noreferrer"
            >
              Facebook
            </a>
          </li>
          <li>
            <a
              href="https://twitter.com"
              target="_blank"
              rel="noopener noreferrer"
            >
              Twitter
            </a>
          </li>
          <li>
            <a
              href="https://instagram.com"
              target="_blank"
              rel="noopener noreferrer"
            >
              Instagram
            </a>
          </li>
        </ul>
        <div className="footer-copyright">
          &copy; 2024 CreeperKeeper. All rights reserved.
        </div>
      </div>
    </footer>
  );
}
