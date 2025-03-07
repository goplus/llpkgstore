import clipboardImg from '../../assets/clipboard.svg';
import githubImg from '../../assets/github.svg';
import conanImg from '../../assets/conan.svg';
import rollbackImg from '../../assets/rollback.svg';
import { CSSTransition } from 'react-transition-group';
import { useRef } from 'react';

interface TitleProps {
    name?: string;
    version: string;
    setVersion: (tag: string) => void;
}

const Title: React.FC<TitleProps> = ({ name, version, setVersion }) => {
    const rollbackRef = useRef<HTMLButtonElement>(null);
    const copyVersionToClipboard = () => {
        navigator.clipboard.writeText(`llgo get ${name}@${version}`);
    };
    return (
        <div className="flex h-1/3 flex-col items-start gap-4 border-b border-gray-300 px-4 pb-4 sm:h-1/6 sm:flex-row">
            <div className="flex flex-col text-left">
                <h1 className="text-3xl font-bold">{name}</h1>
                <a
                    className="text-xs font-light text-gray-400"
                    href={`https://github.com/goplus/llpkg/${name}`}
                    target="_blank"
                    rel="noreferrer"
                >
                    github.com/goplus/llpkg/{name}
                </a>
            </div>
            <div className="ml-auto">
                <div className="flex flex-row overflow-hidden rounded-lg border border-gray-300 pr-2 pl-5">
                    <p className="py-2 text-gray-700">
                        llgo get {name}@{version}
                    </p>
                    <CSSTransition
                        in={version !== 'latest'}
                        timeout={300}
                        classNames={{
                            enter: 'opacity-0',
                            enterActive:
                                'opacity-100 transition-all duration-300',
                            exitActive: 'opacity-0 transition-all duration-300',
                        }}
                        unmountOnExit
                        nodeRef={rollbackRef}
                    >
                        <button
                            className="cursor-pointer pl-3 text-white opacity-0 transition duration-200 hover:scale-110 active:scale-90"
                            onClick={() => setVersion('latest')}
                            ref={rollbackRef}
                        >
                            <img src={rollbackImg} className="h-4 w-4" />
                        </button>
                    </CSSTransition>
                    <button
                        className="cursor-pointer px-3 text-white transition duration-200 hover:scale-110 active:scale-90"
                        onClick={copyVersionToClipboard}
                    >
                        <img src={clipboardImg} className="h-4 w-4" />
                    </button>
                </div>
                <div className="mt-2 text-right">
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
