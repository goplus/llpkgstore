import { useEffect, useState } from 'react';
import { VersionData } from '../../tools/parser/types';
import { versionParser } from '../../tools/parser/parser';
import Modal from '../modal';
import clipboardImg from '../../assets/clipboard.svg';
import githubImg from '../../assets/github.svg';
import conanImg from '../../assets/conan.svg';

interface DetailModalProps {
    modalOpen: boolean;
    setModalOpen: (open: boolean) => void;
    name?: string;
    data?: VersionData;
}

const DetailModal: React.FC<DetailModalProps> = ({
    modalOpen,
    setModalOpen,
    name,
    data,
}) => {
    const [originVersion, setOriginVersion] = useState('');
    const [convertVersion, setConvertVersion] = useState('');
    const [version, setVersion] = useState('latest');
    const copyVersionToClipboard = () => {
        navigator.clipboard.writeText(`llgo get ${name}@${version}`);
    };
    useEffect(() => {
        setVersion('latest');
    }, [name]);
    return (
        <Modal
            isOpen={modalOpen}
            onClose={() => setModalOpen(false)}
            title="Details"
        >
            <div className="flex flex-row items-center gap-4 border-b border-gray-300 px-4 pb-4">
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
            <div>
                {/* <Details /> */}
                {data && name && (
                    <>
                        <div className="my-2 flex flex-row justify-around text-base">
                            <div className="flex items-center gap-2 rounded-3xl border border-gray-300 px-2 py-1">
                                <input
                                    onChange={(e) =>
                                        setOriginVersion(e.target.value)
                                    }
                                    className="px-1 focus-visible:outline-0"
                                    placeholder="Original version"
                                />
                            </div>
                            <div className="flex items-center gap-2 rounded-3xl border border-gray-300 px-2 py-1">
                                <input
                                    onChange={(e) =>
                                        setConvertVersion(e.target.value)
                                    }
                                    className="px-1 focus-visible:outline-0"
                                    placeholder="Converted version"
                                />
                            </div>
                        </div>
                        {versionParser(
                            data[name],
                            originVersion,
                            convertVersion,
                            0,
                            10,
                        ).map((ver, index) => {
                            return (
                                <div
                                    key={index}
                                    className="flex flex-row items-center gap-4"
                                >
                                    <span className="min-w-16 text-left text-lg font-bold">
                                        {ver.original}
                                    </span>
                                    <span>{ver.converted.join(' / ')}</span>
                                </div>
                            );
                        })}
                    </>
                )}
            </div>
        </Modal>
    );
};

export default DetailModal;
