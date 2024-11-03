import { useAuth0 } from "@auth0/auth0-react";

const LogoutButton = () => {
  const { logout } = useAuth0();

  return (
    <div className="bg-[--primary-background] mb-16 w-full flex justify-center">
      <button
        onClick={() => logout({ logoutParams: { returnTo: window.location.origin } })}
        className="mx-0 w-full bg-blue-500 hover:bg-blue-700 text-white py-4 px-6 rounded"
      >
        Log Out
      </button >
    </div>
  );
};

export default LogoutButton;