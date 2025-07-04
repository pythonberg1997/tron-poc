// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.23;

contract POC {
    uint256 public value;
    
    function read() public view returns (uint256) {
        return value;
    }

    function write(uint256 _value) public {
        value = _value;
    }
}