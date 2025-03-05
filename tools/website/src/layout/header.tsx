const Header = () => {
  return (
    <div className="sticky top-0 flex flex-row items-center bg-white/80 px-8 py-2 shadow-sm backdrop-blur-lg">
      <div className="mr-auto flex flex-col text-left">
        <span className="text-3xl font-bold">llpkgstore</span>
        <span className="text-sm text-gray-600">
          Version map searching for llpkg.
        </span>
      </div>
      <a
        href="./llpkgstore.json"
        target="_blank"
        rel="noreferrer"
        className="text-blue-500"
      >
        Raw JSON
      </a>
    </div>
  );
};

export default Header;
