import { useState } from 'react';
import Item from './items';
import { VersionData } from '../../tools/parser/types';
import DetailModal from './detail';

interface ListProps {
  data?: VersionData;
  titles: string[];
}

const List: React.FC<ListProps> = ({ data, titles }) => {
  const [modalOpen, setModalOpen] = useState(false);
  const [name, setDetailName] = useState<string>();
  return (
    <>
      <DetailModal
        modalOpen={modalOpen}
        setModalOpen={setModalOpen}
        data={data}
        name={name}
      />
      <div className="grid grid-cols-2 gap-4 px-32 py-16 text-gray-600">
        {data ? (
          titles.map((key, index) => {
            return (
              <Item
                key={key}
                name={key}
                data={data[key]}
                index={index}
                setInfo={setDetailName}
                setModalOpen={setModalOpen}
              />
            );
          })
        ) : (
          <p>Loading...</p>
        )}
      </div>
    </>
  );
};

export default List;
