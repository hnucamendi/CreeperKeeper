import { useAuth0 } from "@auth0/auth0-react";

const LogoutButton = () => {
  const { logout } = useAuth0();

  return (
    <button
      onClick={() => logout({ logoutParams: { returnTo: window.location.origin } })}
      className="w-full sm:w-1/2 md:w-1/3 lg:w-1/4 bg-blue-500 hover:bg-blue-700 text-white py-4 px-6 rounded"
    >
      Log Out
    </button>
  );
};

export default LogoutButton;