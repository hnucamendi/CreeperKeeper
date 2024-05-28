import { useNavigate } from "react-router-dom";
import Nav from "../components/Nav";
import Footer from "../components/Footer";
import Container from "../components/Container";
import {
  FormContainer,
  Label,
  Select,
  Option,
  Button,
} from "../components/Form/index";

const Home = () => {
  const navigate = useNavigate();
  const handleLogout = () => {
    sessionStorage.clear();
    navigate("/login");
  };

  return (
    <Container className="server-menu">
      <Nav
        listItems={[
          { item: "Home", link: "/home", callback: null },
          { item: "Sign Out", link: "/login", callback: handleLogout },
        ]}
      ></Nav>
      <FormContainer>
        <Label htmlFor="server-dropdown">Select Version</Label>
        <Select id="server-dropdown" />
        <Button>Create Server</Button>
      </FormContainer>
      <FormContainer>
        <Label htmlFor="server-dropdown">Select Saved Instance</Label>
        <Select id="server-dropdown">
          <Option>Test</Option>
          <Option>Test</Option>
          <Option>Test</Option>
          <Option>Test</Option>
          <Option>Test</Option>
          <Option>Test</Option>
        </Select>
        <Button>Startup Server</Button>
      </FormContainer>
      <Footer />
    </Container>
  );
};

export default Home;
