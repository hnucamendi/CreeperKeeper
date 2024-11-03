import { useAuth0 } from "@auth0/auth0-react";

const LoginButton = () => {
  const { loginWithRedirect } = useAuth0();

  return (
    <div className="bg-[--primary-background] mb-16 w-full flex justify-center">
      <button
        onClick={() => loginWithRedirect()}
        className="mx-0 w-full bg-blue-500 hover:bg-blue-700 text-white py-4 px-6 rounded"
      >
        Log In
      </button>
    </div>
  );
};

export default LoginButton;
