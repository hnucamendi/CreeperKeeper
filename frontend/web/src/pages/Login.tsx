import LoginButton from '../component/Login';

export default function Login() {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-75 flex justify-center items-center z-50">
      <div className="bg-white p-6 rounded-lg w-[90%] sm:w-3/4 md:w-1/2 lg:w-1/3 mx-4">
        <h2 className="text-xl sm:text-2xl font-bold mb-4 text-center">Login</h2>
        <div className="flex justify-center">
          <LoginButton />
        </div>
      </div>
    </div>
  );
}