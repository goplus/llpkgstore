import clipboardImg from '../../assets/clipboard.svg';
import githubImg from '../../assets/github.svg';
import conanImg from '../../assets/conan.svg';

interface TitleProps {
    name?: string;
    version: string;
}

const Title: React.FC<TitleProps> = ({ name, version }) => {
    const copyVersionToClipboard = () => {
        navigator.clipboard.writeText(`llgo get ${name}@${version}`);
    };
    return (
        <div className="flex h-1/3 flex-col items-start gap-4 border-b border-gray-300 px-4 pb-4 sm:h-1/6 sm:flex-row">
            <div className="flex flex-col text-left">
                <h1 className="text-3xl font-bold">{name}</h1>
                <span className="text-xs font-light text-gray-400">
                    github.com/goplus/llpkg/{name}
                </span>
            </div>
            <div className="ml-auto">
                <div className="flex flex-row overflow-hidden rounded-lg border border-gray-300 pl-5">
                    <p className="py-2 text-gray-700">
                        llgo get {name}@{version}
                    </p>
                    <button
                        className="cursor-pointer px-5 text-white transition duration-200 hover:scale-110 active:scale-90"
                        onClick={copyVersionToClipboard}
                    >
                        <img src={clipboardImg} className="h-4 w-4" />
                    </button>
                </div>
                <div className="text-right">
                    <a
                        href={`https://github.com/goplus/llpkg/tree/main/${name}`}
                        target="_blank"
                        rel="noreferrer"
                    >
                        <img
                            className="mx-1 inline-block h-4 w-4"
                            src={githubImg}
                        />
                    </a>
                    <a
                        href={`https://conan.io/center/recipes/${name}`}
                        target="_blank"
                        rel="noreferrer"
                    >
                        <img
                            className="mx-1 inline-block h-4 w-4"
                            src={conanImg}
                        />
                    </a>
                </div>
            </div>
        </div>
    );
};

export default Title;
