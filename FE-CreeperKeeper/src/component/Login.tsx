import { useAuth0 } from "@auth0/auth0-react";

const LoginButton = () => {
  const { loginWithRedirect } = useAuth0();

  return (
    <button
      onClick={() => loginWithRedirect()}
      className="w-full bg-blue-500 hover:bg-blue-700 text-white py-4 px-6 rounded"
    >
      Log In
    </button>
  );
};

export default LoginButton;
