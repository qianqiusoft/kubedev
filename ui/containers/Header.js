import React, { useState, useMemo } from 'react';
import styled from '@emotion/styled';
import { useWorker } from 'react-hooks-worker';
import Downshift from 'downshift';
import Fuse from 'fuse.js';

import { primaryDark, fontColorWhite } from '../util/colors';
import logo from '../assets/logo.svg';
import Input from '../components/Input';
import { formatSearchResponse } from '../state-management/general-managements';
import { navigate } from '@reach/router';

const HeaderContainer = styled.div`
  display: flex;
  align-items: center;
  position: relative;
  height: 60px;
  padding: 16px;
  background: ${props => props.theme.header};
  color: white;
  box-shadow: 0 0 0.4rem rgba(0, 0, 0, 0.1), 0 0.1rem 0.8rem rgba(0, 0, 0, 0.2);
  overflow: inherit;
`;

const LogoContainer = styled.div`
  display: flex;
  align-items: center;
  width: 204px;
  height: 100%;
`;

const Image = styled.img`
  height: 100%;
`;

const Title = styled.h1`
  margin-left: 16px;
  font-size: 24px;
`;

const InputContainer = styled.div`
  position: relative;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  flex: 1;
  border-radius: 3px;
  background: ${props => props.theme.background};

  box-shadow: ${({ focus }) =>
    focus ? '0px 2px 2px rgba(0, 0, 0, 0.25)' : null};

  input {
    width: 500px;
    height: 40px;
    padding: 16px;
    border: none;
    background: transparent;
    color: ${props => props.theme.sidebarFontColor};
    outline: none;
  }

  svg {
    margin-left: 7px;
    fill: ${props => props.theme.sidebarFontColor};
    opacity: 0.54;
  }
`;

const SearchIcon = styled.button`
  background: none;
  border: none;
`;

const AutoComplete = styled.div`
  position: absolute;
  width: 100%;
  top: 38px;
  background: ${props => props.theme.background};
  color: ${props => props.theme.sidebarFontColor};
  border-bottom-left-radius: 3px;
  border-bottom-right-radius: 3px;
  box-shadow: 0px 2px 2px rgba(0, 0, 0, 0.25);
  overflow: hidden;
`;

const SearchItem = styled.div`
  display: flex;
  flex-direction: column;
  padding: 10px;
  background: ${props =>
    props.highlighted ? primaryDark : props.theme.background};
  color: ${props =>
    props.highlighted ? fontColorWhite : props.theme.sidebarFontColor};
  font-size: 14px;

  span {
    font-size: 12px;
  }
`;

const worker = new Worker('../workers/search.js');

export default function Header() {
  const [searchDate, setSearchDate] = useState(new Date());
  const [focus, setFocus] = useState(false);
  const { result, error } = useWorker(worker, searchDate);

  const handleFocus = () => {
    setFocus(true);
    setSearchDate(new Date());
  };

  const handleBlur = () => {
    setFocus(false);
  };

  const items = useMemo(() => formatSearchResponse(result), [result]);

  let fuse = new Fuse(items, {
    keys: ['type', 'namespace', 'name']
  });

  return (
    <HeaderContainer>
      <LogoContainer>
        <Image src={logo} alt="KubeDev logo" />
        <Title>KubeDev</Title>
      </LogoContainer>
      <Downshift
        onChange={selection =>
          navigate(
            `/${selection.namespace}/${selection.type}/${selection.name}/info`
          )
        }
        itemToString={item => (item ? item.name : '')}
      >
        {({
          getInputProps,
          getItemProps,
          getLabelProps,
          isOpen,
          inputValue,
          highlightedIndex,
          selectedItem
        }) => (
          <div>
            <InputContainer focus={focus}>
              <SearchIcon {...getLabelProps({ 'aria-label': 'search' })}>
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                >
                  <path d="M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z" />
                  <path d="M0 0h24v24H0z" fill="none" />
                </svg>
              </SearchIcon>
              <Input
                {...getInputProps({
                  placeholder: 'Search',
                  onFocus: handleFocus,
                  onBlur: handleBlur
                })}
              />
              {isOpen ? (
                <AutoComplete>
                  {fuse.search(inputValue, { limit: 10 }).map((item, index) => (
                    <SearchItem
                      {...getItemProps({
                        key: `${item.type}-${item.namespace}-${item.name}`,
                        index,
                        item,
                        highlighted: highlightedIndex === index,
                        selected: selectedItem === item
                      })}
                    >
                      <strong>{item.name}</strong>
                      <span>
                        {item.namespace} - {item.type}
                      </span>
                    </SearchItem>
                  ))}
                </AutoComplete>
              ) : null}
            </InputContainer>
          </div>
        )}
      </Downshift>
    </HeaderContainer>
  );
}
